package baud

// ConfirmInputProps configures the destructive-action guard
// (design/README.md "Feedback"): "type <name> to confirm" with the danger
// action Btn disabled until the typed value exactly matches.
type ConfirmInputProps struct {
	// Expect is the exact value the user must type (the resource name).
	Expect string
	// ActionLabel labels the danger Btn (author lowercase per type rules).
	ActionLabel string
	// Name is the input's form field name (default "confirm") — the typed
	// value submits with the enclosing form.
	Name string
	// ID names the input element (label/test hook).
	ID string
}

func (p ConfirmInputProps) name() string {
	return or(p.Name, "confirm")
}

// confirmScript is the inline hyperscript guard on the input — purely-local
// UI, deliberately NOT a baud._hs behavior (one consumer, one element). On
// every input event it compares the typed value with the element's own
// data-confirm attribute: an exact match enables the danger Btn; any other
// non-empty value shows the err border while typing; empty clears the err
// state but keeps the Btn disabled.
const confirmScript = `on input
set wrap to the closest <.input-wrap/>
set btn to the first <button.btn/> in the closest <.confirm/>
if my value is @data-confirm
  remove @disabled from btn
  remove .err from wrap
  call me.removeAttribute('aria-invalid')
else
  add @disabled to btn
  if my value is ''
    remove .err from wrap
    call me.removeAttribute('aria-invalid')
  else
    add .err to wrap
    call me.setAttribute('aria-invalid', 'true')
  end
end`
