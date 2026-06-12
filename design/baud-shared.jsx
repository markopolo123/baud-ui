// baud/ui — shared primitives (buttons, inputs, badges, tabs, panels)
const { useState, useEffect, useRef, useMemo, useCallback } = React;

function Kbd({ children }) { return <span className="kbd">{children}</span>; }

function Btn({ variant, glyph, kbd, active, children, ...rest }) {
  const cls = ['btn'];
  if (variant === 'primary') cls.push('btn-primary');
  if (variant === 'danger') cls.push('btn-danger');
  if (variant === 'ghost') cls.push('btn-ghost');
  if (active) cls.push('is-active');
  return (
    <button className={cls.join(' ')} {...rest}>
      {glyph ? <span className="btn-glyph">{glyph}</span> : null}
      {children}
      {kbd ? <Kbd>{kbd}</Kbd> : null}
    </button>
  );
}

function BtnGroup({ children }) { return <span className="btn-group">{children}</span>; }

function Badge({ tone = 'neutral', variant = 'tint', dot, children }) {
  const v = variant === 'solid' ? 'bd-solid' : variant === 'outline' ? 'bd-outline' : 'bd-tint';
  return (
    <span className={`badge ${v} tone-${tone}`}>
      {dot ? <span className="badge-dot"></span> : null}
      {children}
    </span>
  );
}

function Dot({ tone = 'ok', pulse }) {
  return <span className={`dot tone-${tone}${pulse ? ' pulse' : ''}`}></span>;
}

function Field({ label, hint, error, children }) {
  return (
    <div className="field">
      {label ? <span className="field-label">{label}</span> : null}
      {children}
      {error ? <span className="field-hint err">✗ {error}</span> :
        hint ? <span className="field-hint">{hint}</span> : null}
    </div>
  );
}

function Input({ prefix, suffix, error, style, ...rest }) {
  return (
    <span className={`input-wrap${error ? ' err' : ''}`} style={style}>
      {prefix ? <span className="affix">{prefix}</span> : null}
      <input className="input" {...rest} />
      {suffix ? <span className="affix">{suffix}</span> : null}
    </span>
  );
}

// Click-outside helper
function useClickOutside(ref, onOut, active) {
  useEffect(() => {
    if (!active) return;
    const h = (e) => { if (ref.current && !ref.current.contains(e.target)) onOut(); };
    document.addEventListener('mousedown', h);
    return () => document.removeEventListener('mousedown', h);
  }, [active, onOut]);
}

function Select({ value, options, onChange, width, alignRight }) {
  const [open, setOpen] = useState(false);
  const ref = useRef(null);
  useClickOutside(ref, () => setOpen(false), open);
  const current = options.find((o) => o.value === value);
  return (
    <span className="select" ref={ref} style={width ? { width } : null}>
      <button className="select-trigger" onClick={() => setOpen(!open)} style={width ? { width } : null}>
        <span>{current ? current.label : '—'}</span>
        <span className="select-chev">{open ? '▲' : '▼'}</span>
      </button>
      {open ? (
        <div className={`menu${alignRight ? ' menu-right' : ''}`}>
          {options.map((o) =>
            o.sep ? <div className="menu-sep" key={o.key || 'sep'}></div> : (
              <button
                key={o.value}
                className={`menu-item${o.value === value ? ' is-active' : ''}`}
                onClick={() => { onChange(o.value); setOpen(false); }}
              >
                <span>{o.label}</span>
                {o.value === value ? <span className="menu-check">✓</span> : null}
              </button>
            )
          )}
        </div>
      ) : null}
    </span>
  );
}

function Checkbox({ checked, onChange, disabled, children }) {
  return (
    <span
      className={`cbx${checked ? ' on' : ''}${disabled ? ' disabled' : ''}`}
      onClick={disabled ? null : () => onChange(!checked)}
    >
      <span className="cbx-box">{checked ? '[x]' : '[ ]'}</span>
      <span>{children}</span>
    </span>
  );
}

function Radio({ checked, onChange, children }) {
  return (
    <span className={`cbx${checked ? ' on' : ''}`} onClick={() => onChange()}>
      <span className="cbx-box">{checked ? '(•)' : '( )'}</span>
      <span>{children}</span>
    </span>
  );
}

function Toggle({ options, value, onChange }) {
  return (
    <span className="tg">
      {options.map((o) => (
        <span key={o} className={`tg-opt${o === value ? ' sel' : ''}`} onClick={() => onChange(o)} style={{ cursor: 'pointer' }}>
          {o}
        </span>
      ))}
    </span>
  );
}

function Tabs({ items, value, onChange, boxed }) {
  return (
    <div className={`tabs ${boxed ? 'tabs-boxed' : 'tabs-underline'}`}>
      {items.map((t) => (
        <button key={t.id} className={`tab${t.id === value ? ' is-active' : ''}`} onClick={() => onChange(t.id)}>
          {t.label}
          {t.badge != null ? <span className="tab-badge">{t.badge}</span> : null}
        </button>
      ))}
    </div>
  );
}

function Panel({ title, acts, children, style, bodyStyle }) {
  return (
    <div className="panel" style={style}>
      <div className="panel-hd">
        <span className="panel-title">{title}</span>
        {acts ? <span className="panel-acts">{acts}</span> : null}
      </div>
      <div className="panel-bd" style={bodyStyle}>{children}</div>
    </div>
  );
}

Object.assign(window, {
  Kbd, Btn, BtnGroup, Badge, Dot, Field, Input, Select, Checkbox, Radio,
  Toggle, Tabs, Panel, useClickOutside,
});
