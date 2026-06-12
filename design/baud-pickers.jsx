// baud/ui — pickers: Combobox, DatePicker
const { useState: useStateP, useRef: useRefP, useMemo: useMemoP, useEffect: useEffectP } = React;

/* ---------------- Combobox ---------------- */
// options: [{value, label, meta?}]
function Combobox({ options, value, onChange, placeholder, width }) {
  const [open, setOpen] = useStateP(false);
  const [q, setQ] = useStateP('');
  const [hl, setHl] = useStateP(0);
  const ref = useRefP(null);
  useClickOutside(ref, () => { setOpen(false); setQ(''); }, open);

  const current = options.find((o) => o.value === value);
  const filtered = useMemoP(() => {
    const t = q.trim().toLowerCase();
    if (!t) return options;
    return options.filter((o) => (o.label + ' ' + (o.meta || '')).toLowerCase().includes(t));
  }, [q, options]);
  useEffectP(() => { setHl(0); }, [q]);

  const pick = (o) => { onChange(o.value); setOpen(false); setQ(''); };
  const key = (e) => {
    if (e.key === 'Escape') { setOpen(false); setQ(''); }
    else if (e.key === 'ArrowDown') { e.preventDefault(); setHl((h) => Math.min(h + 1, filtered.length - 1)); }
    else if (e.key === 'ArrowUp') { e.preventDefault(); setHl((h) => Math.max(h - 1, 0)); }
    else if (e.key === 'Enter' && filtered[hl]) { e.preventDefault(); pick(filtered[hl]); }
  };

  const mark = (label) => {
    const t = q.trim().toLowerCase();
    if (!t) return label;
    const i = label.toLowerCase().indexOf(t);
    if (i < 0) return label;
    return (
      <React.Fragment>
        {label.slice(0, i)}
        <span className="cb-match">{label.slice(i, i + t.length)}</span>
        {label.slice(i + t.length)}
      </React.Fragment>
    );
  };

  return (
    <span className="select" ref={ref} style={width ? { width } : null}>
      <span className={`input-wrap${open ? '' : ''}`} style={{ width: '100%', cursor: 'text' }}
            onClick={() => setOpen(true)}>
        <span className="affix">⌕</span>
        <input
          className="input"
          placeholder={current && !open ? undefined : (placeholder || 'search…')}
          value={open ? q : (current ? current.label : '')}
          onChange={(e) => { setQ(e.target.value); setOpen(true); }}
          onFocus={() => setOpen(true)}
          onKeyDown={key}
        />
        <span className="select-chev">{open ? '▲' : '▼'}</span>
      </span>
      {open ? (
        <div className="menu" style={{ width: '100%' }}>
          {filtered.length === 0 ? <div className="palette-empty">no match for “{q}”</div> : null}
          {filtered.map((o, i) => (
            <button key={o.value}
                    className={`menu-item${i === hl ? ' hl-row' : ''}${o.value === value ? ' is-active' : ''}`}
                    onMouseEnter={() => setHl(i)}
                    onClick={() => pick(o)}>
              <span>{mark(o.label)}</span>
              <span style={{ display: 'inline-flex', gap: 8, alignItems: 'center' }}>
                {o.meta ? <span className="tone-faint" style={{ fontSize: 'var(--fs-sm)' }}>{o.meta}</span> : null}
                {o.value === value ? <span className="menu-check">✓</span> : null}
              </span>
            </button>
          ))}
        </div>
      ) : null}
    </span>
  );
}

/* ---------------- DatePicker ---------------- */
const DP_MONTHS = ['Jan','Feb','Mar','Apr','May','Jun','Jul','Aug','Sep','Oct','Nov','Dec'];
const DP_DOW = ['Mo','Tu','We','Th','Fr','Sa','Su'];

function dpFmt(d) {
  if (!d) return '';
  const p = (n) => String(n).padStart(2, '0');
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())}`;
}
function dpSame(a, b) {
  return a && b && a.getFullYear() === b.getFullYear() && a.getMonth() === b.getMonth() && a.getDate() === b.getDate();
}

function DatePicker({ value, onChange, width, presets = true }) {
  const [open, setOpen] = useStateP(false);
  const today = useMemoP(() => new Date(), []);
  const [view, setView] = useStateP(() => {
    const d = value || new Date();
    return { y: d.getFullYear(), m: d.getMonth() };
  });
  const ref = useRefP(null);
  useClickOutside(ref, () => setOpen(false), open);

  const grid = useMemoP(() => {
    const first = new Date(view.y, view.m, 1);
    const offset = (first.getDay() + 6) % 7; // monday-first
    const start = new Date(view.y, view.m, 1 - offset);
    return Array.from({ length: 42 }, (_, i) =>
      new Date(start.getFullYear(), start.getMonth(), start.getDate() + i));
  }, [view]);

  const nav = (dm) => setView(({ y, m }) => {
    const d = new Date(y, m + dm, 1);
    return { y: d.getFullYear(), m: d.getMonth() };
  });
  const pick = (d) => { onChange(d); setOpen(false); };
  const preset = (days) => {
    const d = new Date(today.getFullYear(), today.getMonth(), today.getDate() - days);
    setView({ y: d.getFullYear(), m: d.getMonth() });
    pick(d);
  };

  return (
    <span className="select" ref={ref} style={width ? { width } : null}>
      <button className="select-trigger" onClick={() => setOpen(!open)} style={width ? { width } : null}>
        <span style={{ display: 'inline-flex', gap: 6, alignItems: 'center' }}>
          <span className="affix" style={{ color: 'var(--fg-faint)' }}>▦</span>
          <span style={{ fontVariantNumeric: 'tabular-nums' }}>{value ? dpFmt(value) : 'pick date'}</span>
        </span>
        <span className="select-chev">{open ? '▲' : '▼'}</span>
      </button>
      {open ? (
        <div className="menu dp" style={{ padding: 6 }}>
          <div className="dp-hd">
            <button className="x-btn" onClick={() => nav(-12)}>«</button>
            <button className="x-btn" onClick={() => nav(-1)}>‹</button>
            <span className="dp-title">{DP_MONTHS[view.m]} {view.y}</span>
            <button className="x-btn" onClick={() => nav(1)}>›</button>
            <button className="x-btn" onClick={() => nav(12)}>»</button>
          </div>
          <div className="dp-grid">
            {DP_DOW.map((d) => <span key={d} className="dp-dow">{d}</span>)}
            {grid.map((d, i) => {
              const out = d.getMonth() !== view.m;
              const cls = ['dp-day'];
              if (out) cls.push('out');
              if (dpSame(d, today)) cls.push('today');
              if (dpSame(d, value)) cls.push('sel');
              return (
                <button key={i} className={cls.join(' ')} onClick={() => pick(d)}>
                  {d.getDate()}
                </button>
              );
            })}
          </div>
          {presets ? (
            <div className="dp-presets">
              <button className="dp-preset" onClick={() => preset(0)}>today</button>
              <button className="dp-preset" onClick={() => preset(1)}>-1d</button>
              <button className="dp-preset" onClick={() => preset(7)}>-7d</button>
              <button className="dp-preset" onClick={() => preset(30)}>-30d</button>
            </div>
          ) : null}
        </div>
      ) : null}
    </span>
  );
}

Object.assign(window, { Combobox, DatePicker });
