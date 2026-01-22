package components

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type FormModal struct {
	*tview.Box
	form   *tview.Form
	width  int
	height int
}

func NewFormModal(form *tview.Form, width, height int) *FormModal {
	return &FormModal{
		Box:    tview.NewBox(),
		form:   form,
		width:  width,
		height: height,
	}
}

func (m *FormModal) Draw(screen tcell.Screen) {
	screenWidth, screenHeight := screen.Size()

	width := m.width
	height := m.height
	if height == 0 {
		height = m.form.GetFormItemCount()*2 + 5
	}

	x := (screenWidth - width) / 2
	y := (screenHeight - height) / 2

	for row := 0; row < screenHeight; row++ {
		for col := 0; col < screenWidth; col++ {
			screen.SetContent(col, row, ' ', nil, tcell.StyleDefault.Background(tcell.ColorBlack))
		}
	}

	m.form.SetRect(x, y, width, height)
	m.form.Draw(screen)
}

func (m *FormModal) Focus(delegate func(p tview.Primitive)) {
	delegate(m.form)
}

func (m *FormModal) HasFocus() bool {
	return m.form.HasFocus()
}

func (m *FormModal) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return m.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		if m.form.HasFocus() {
			if handler := m.form.InputHandler(); handler != nil {
				handler(event, setFocus)
			}
		}
	})
}

func (m *FormModal) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return m.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		if m.form.InRect(event.Position()) {
			if handler := m.form.MouseHandler(); handler != nil {
				return handler(action, event, setFocus)
			}
		}
		return false, nil
	})
}

type FieldType int

const (
	FieldTypeInput FieldType = iota
	FieldTypePassword
	FieldTypeDropDown
)

type FormField struct {
	Type         FieldType
	Label        string
	InitialValue string
	Options      []string
	InitialIndex int
	FieldWidth   int
	OnChanged    func(string)
	OnSelected   func(string, int)
}

type FormDialog struct {
	pages         *tview.Pages
	app           *tview.Application
	title         string
	fields        []FormField
	submitLabel   string
	cancelLabel   string
	onSubmit      func(values map[string]string) error
	onCancel      func()
	pageName      string
	modalWidth    int
	modalHeight   int
	form          *tview.Form
	modal         *FormModal
	inputValues   map[string]string
	escapeToClose bool
}

type FormDialogConfig struct {
	Title         string
	Fields        []FormField
	SubmitLabel   string
	CancelLabel   string
	OnSubmit      func(values map[string]string) error
	OnCancel      func()
	PageName      string
	ModalWidth    int
	ModalHeight   int
	EscapeToClose bool
}

func NewFormDialog(pages *tview.Pages, app *tview.Application, config FormDialogConfig) *FormDialog {
	if config.SubmitLabel == "" {
		config.SubmitLabel = "Submit"
	}
	if config.CancelLabel == "" {
		config.CancelLabel = "Cancel"
	}
	if config.PageName == "" {
		config.PageName = "form-dialog"
	}
	if config.ModalWidth == 0 {
		config.ModalWidth = 60
	}
	if config.ModalHeight == 0 {
		config.ModalHeight = 0
	}

	fd := &FormDialog{
		pages:         pages,
		app:           app,
		title:         config.Title,
		fields:        config.Fields,
		submitLabel:   config.SubmitLabel,
		cancelLabel:   config.CancelLabel,
		onSubmit:      config.OnSubmit,
		onCancel:      config.OnCancel,
		pageName:      config.PageName,
		modalWidth:    config.ModalWidth,
		modalHeight:   config.ModalHeight,
		inputValues:   make(map[string]string),
		escapeToClose: config.EscapeToClose,
	}

	fd.buildForm()
	return fd
}

func (fd *FormDialog) buildForm() {
	fd.form = tview.NewForm()

	for _, field := range fd.fields {
		fd.addField(field)
	}

	fd.addButtons()
	fd.configureFormAppearance()
	fd.configureEscapeHandler()
}

func (fd *FormDialog) addField(field FormField) {
	switch field.Type {
	case FieldTypeInput:
		fd.addInputField(field)
	case FieldTypePassword:
		fd.addPasswordField(field)
	case FieldTypeDropDown:
		fd.addDropDownField(field)
	}
}

func (fd *FormDialog) addInputField(field FormField) {
	width := fd.getFieldWidth(field.FieldWidth)
	fd.inputValues[field.Label] = field.InitialValue
	fd.form.AddInputField(field.Label, field.InitialValue, width, nil, fd.createChangeHandler(field.Label, field.OnChanged))
}

func (fd *FormDialog) addPasswordField(field FormField) {
	width := fd.getFieldWidth(field.FieldWidth)
	passwordField := tview.NewInputField().
		SetLabel(field.Label).
		SetFieldWidth(width).
		SetMaskCharacter('*').
		SetChangedFunc(fd.createChangeHandler(field.Label, field.OnChanged))
	fd.form.AddFormItem(passwordField)
	fd.inputValues[field.Label] = field.InitialValue
}

func (fd *FormDialog) addDropDownField(field FormField) {
	initialIndex := fd.getInitialIndex(field.InitialIndex)
	if initialIndex < len(field.Options) {
		fd.inputValues[field.Label] = field.Options[initialIndex]
	}
	fd.form.AddDropDown(field.Label, field.Options, initialIndex, fd.createSelectHandler(field.Label, field.OnSelected))
}

func (fd *FormDialog) getFieldWidth(width int) int {
	if width == 0 {
		return 30
	}
	return width
}

func (fd *FormDialog) getInitialIndex(index int) int {
	if index < 0 {
		return 0
	}
	return index
}

func (fd *FormDialog) createChangeHandler(label string, onChange func(string)) func(string) {
	return func(text string) {
		fd.inputValues[label] = text
		if onChange != nil {
			onChange(text)
		}
	}
}

func (fd *FormDialog) createSelectHandler(label string, onSelected func(string, int)) func(string, int) {
	return func(option string, index int) {
		fd.inputValues[label] = option
		if onSelected != nil {
			onSelected(option, index)
		}
	}
}

func (fd *FormDialog) addButtons() {
	fd.form.AddButton(fd.submitLabel, func() {
		if fd.onSubmit != nil {
			if err := fd.onSubmit(fd.inputValues); err != nil {
				ShowError(fd.pages, fd.app, err)
				return
			}
		}
		fd.Close()
	})

	fd.form.AddButton(fd.cancelLabel, func() {
		if fd.onCancel != nil {
			fd.onCancel()
		}
		fd.Close()
	})
}

func (fd *FormDialog) configureFormAppearance() {
	fd.form.SetBorder(true).
		SetTitle(fd.title).
		SetTitleAlign(tview.AlignCenter)
}

func (fd *FormDialog) configureEscapeHandler() {
	if fd.escapeToClose {
		fd.form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEsc {
				if fd.onCancel != nil {
					fd.onCancel()
				}
				fd.Close()
				return nil
			}
			return event
		})
	}
}

func (fd *FormDialog) Show() {
	fd.modal = NewFormModal(fd.form, fd.modalWidth, fd.modalHeight)
	fd.pages.AddPage(fd.pageName, fd.modal, true, true)
	fd.app.SetFocus(fd.modal)
}

func (fd *FormDialog) Close() {
	fd.pages.RemovePage(fd.pageName)
}

func (fd *FormDialog) GetForm() *tview.Form {
	return fd.form
}

func (fd *FormDialog) AddField(field FormField) {
	fd.fields = append(fd.fields, field)
	fd.buildForm()
}

func (fd *FormDialog) RemoveField(label string) {
	for i, field := range fd.fields {
		if field.Label == label {
			fd.fields = append(fd.fields[:i], fd.fields[i+1:]...)
			break
		}
	}
	fd.buildForm()
}

func (fd *FormDialog) GetValue(label string) string {
	return fd.inputValues[label]
}

func (fd *FormDialog) SetValue(label, value string) {
	fd.inputValues[label] = value
}
