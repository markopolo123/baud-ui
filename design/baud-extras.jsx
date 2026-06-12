// baud/ui — extras: TagInput, Progress, Spinner, DefList, ConfirmInput,
// Breadcrumb, Tooltip, PanelState, Pagination, SplitPane, DiffViewer
const { useState: useStateX, useRef: useRefX, useMemo: useMemoX, useEffect: useEffectX } = React;

/* ---------------- TagInput / multi-select ---------------- */
// value: ['key=val', ...]; suggestions: ['region=use1', ...]
function TagInput({ value, onChange, suggestions = [], placeholder }) {
  const [q, setQ] = useStateX('');
  const [open, setOpen] = useStateX(false);
  const ref = useRefX(null);
  useClickOutside(ref, () => setOpen(false), open);
  const avail = useMemoX(() =>
    suggestions.filter((s) => !value.includes(s) && (!q || s.includes(q.toLowerCase()))),
    [suggestions, value, q]);
  const add = (t) => { if (t && !value.includes(t)) onChange([...value, t]); setQ(''); };
  const key = (e) => {
    if (e.key === 'Enter' && q.trim()) { e.preventDefault(); add(q.trim()); }
    else if (e.key === 'Backspace' && !q && value.length) onChange(value.slice(0, -1));
    else if (e.key === 'Escape') setOpen(false);
  };
  return (
    <span className="select" ref={ref} style={{ width: '100%' }}>
      <span className="tags-wrap" onClick={() => setOpen(true)}>
        {value.map((t) => {
          const [k, v] = t.includes('=') ? t.split(/=(.*)/) : [null, t];
          return (
            <span className="tag-chip" key={t}>
              {k ? <span className="tag-k">{k}=</span> : null}{v}
              <button className="x-btn" onClick={(e) => { e.stopPropagation(); onChange(value.filter((x) => x !== t)); }}>✕</button>
            </span>
          );
        })}
        <input className="tags-input" value={q} placeholder={value.length ? '' : placeholder}
               onChange={(e) => { setQ(e.target.value); setOpen(true); }}
               onFocus={() => setOpen(true)} onKeyDown={key} />
      </span>
      {open && avail.length ? (
        <div className="menu" style={{ width: '100%' }}>
          {avail.slice(0, 8).map((s) => (
            <button key={s} className="menu-item" onClick={() => add(s)}>
              <span>{s}</span><span className="tone-faint" style={{ fontSize: 'var(--fs-sm)' }}>add</span>
            </button>
          ))}
        </div>
      ) : null}
    </span>
  );
}

/* ---------------- Progress (ASCII) + Spinner ---------------- */
function Progress({ value, max = 100, label, chars = 22, tone }) {
  const pct = Math.max(0, Math.min(1, value / max));
  const full = Math.round(pct * chars);
  const t = tone || (pct >= 1 ? 'ok' : pct > 0.85 ? 'warn' : 'accent');
  return (
    <span className="prog">
      {label ? <span className="prog-label">{label}</span> : null}
      <span className={`prog-bar tone-${t}`}>{'▰'.repeat(full)}<span className="tone-faint">{'▱'.repeat(chars - full)}</span></span>
      <span className="prog-pct">{Math.round(pct * 100)}%</span>
    </span>
  );
}

const SPIN_FRAMES = ['⠋','⠙','⠹','⠸','⠼','⠴','⠦','⠧','⠇','⠏'];
function Spinner({ tone = 'accent' }) {
  const [i, setI] = useStateX(0);
  useEffectX(() => {
    if (window.matchMedia && window.matchMedia('(prefers-reduced-motion: reduce)').matches) return;
    const t = setInterval(() => setI((x) => (x + 1) % SPIN_FRAMES.length), 80);
    return () => clearInterval(t);
  }, []);
  return <span className={`spinner tone-${tone}`}>{SPIN_FRAMES[i]}</span>;
}

/* ---------------- DefList ---------------- */
// rows: [{k, v, tone?}] — v may be a node
function DefList({ rows, lines }) {
  return (
    <div className={`dl${lines ? ' dl-lines' : ''}`}>
      {rows.map((r) => (
        <React.Fragment key={r.k}>
          <span className="dl-k">{r.k}</span>
          <span className={`dl-v${r.tone ? ` tone-${r.tone}` : ''}`}>{r.v}</span>
        </React.Fragment>
      ))}
    </div>
  );
}

/* ---------------- ConfirmInput ---------------- */
function ConfirmInput({ expect, actionLabel, onConfirm }) {
  const [v, setV] = useStateX('');
  const ok = v === expect;
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
      <span className="field-hint">type <span className="tone-err" style={{ userSelect: 'all' }}>{expect}</span> to confirm</span>
      <div style={{ display: 'flex', gap: 6 }}>
        <Input value={v} onChange={(e) => setV(e.target.value)} placeholder={expect}
               style={{ flex: 1 }} error={v.length > 0 && !ok} />
        <Btn variant="danger" disabled={!ok} onClick={() => { onConfirm(); setV(''); }}>{actionLabel}</Btn>
      </div>
    </div>
  );
}

/* ---------------- Breadcrumb ---------------- */
// items: [{id, label}]
function Breadcrumb({ items, onNav }) {
  return (
    <nav className="crumbs">
      {items.map((it, i) => (
        <React.Fragment key={it.id}>
          {i > 0 ? <span className="crumb-sep">›</span> : null}
          {i === items.length - 1
            ? <span className="crumb cur">{it.label}</span>
            : <button className="crumb" onClick={() => onNav && onNav(it)}>{it.label}</button>}
        </React.Fragment>
      ))}
    </nav>
  );
}

/* ---------------- Tooltip ---------------- */
function Tip({ tip, under, children }) {
  return <span className={`tip${under ? ' tip-under' : ''}`} data-tip={tip}>{children}</span>;
}

/* ---------------- PanelState ---------------- */
function PanelState({ kind = 'empty', title, sub, action }) {
  if (kind === 'skeleton') {
    const widths = [[28, 120, 56, 200], [28, 96, 56, 240], [28, 140, 56, 160], [28, 110, 56, 220], [28, 88, 56, 180]];
    return (
      <div aria-hidden="true">
        {widths.map((row, i) => (
          <div className="skel-row" key={i}>
            {row.map((w, j) => <span className="skel-cell" key={j} style={{ width: w }}></span>)}
          </div>
        ))}
      </div>
    );
  }
  const glyph = kind === 'err' ? '✗' : kind === 'loading' ? null : '∅';
  return (
    <div className={`pstate${kind === 'err' ? ' err' : ''}`}>
      <span className="pstate-glyph">{kind === 'loading' ? <Spinner /> : glyph}</span>
      <span className="pstate-title">{title}</span>
      {sub ? <span className="pstate-sub">{sub}</span> : null}
      {action ? <span style={{ marginTop: 4 }}>{action}</span> : null}
    </div>
  );
}

/* ---------------- Pagination ---------------- */
function Pagination({ page, pageSize, total, onPage, onMore }) {
  const from = (page - 1) * pageSize + 1;
  const to = Math.min(page * pageSize, total);
  const last = Math.ceil(total / pageSize);
  return (
    <div className="pager">
      <span className="pager-info">rows {from.toLocaleString()}–{to.toLocaleString()} of {total.toLocaleString()}</span>
      <span className="pager-spring"></span>
      {onMore ? <button className="pager-btn more" onClick={onMore}>load more ↓</button> : null}
      <button className="pager-btn" disabled={page <= 1} onClick={() => onPage(1)}>|‹</button>
      <button className="pager-btn" disabled={page <= 1} onClick={() => onPage(page - 1)}>‹ prev</button>
      <span>{page}/{last}</span>
      <button className="pager-btn" disabled={page >= last} onClick={() => onPage(page + 1)}>next ›</button>
      <button className="pager-btn" disabled={page >= last} onClick={() => onPage(last)}>›|</button>
    </div>
  );
}

/* ---------------- SplitPane ---------------- */
function SplitPane({ left, right, initial = 50, min = 15, max = 85 }) {
  const [pct, setPct] = useStateX(initial);
  const [drag, setDrag] = useStateX(false);
  const ref = useRefX(null);
  const down = (e) => {
    e.preventDefault();
    setDrag(true);
    const move = (ev) => {
      const r = ref.current.getBoundingClientRect();
      setPct(Math.max(min, Math.min(max, ((ev.clientX - r.left) / r.width) * 100)));
    };
    const up = () => { setDrag(false); window.removeEventListener('pointermove', move); window.removeEventListener('pointerup', up); };
    window.addEventListener('pointermove', move);
    window.addEventListener('pointerup', up);
  };
  return (
    <div className="split" ref={ref} style={{ width: '100%', height: '100%' }}>
      <div className="split-pane" style={{ width: `calc(${pct}% - 4px)` }}>{left}</div>
      <div className={`split-gutter${drag ? ' drag' : ''}`} onPointerDown={down}></div>
      <div className="split-pane" style={{ flex: 1 }}>{right}</div>
    </div>
  );
}

/* ---------------- DiffViewer ---------------- */
// lines: [{t:'ctx'|'add'|'del'|'hunk', a?, b?, text}]
function DiffViewer({ file, lines }) {
  const adds = lines.filter((l) => l.t === 'add').length;
  const dels = lines.filter((l) => l.t === 'del').length;
  return (
    <div className="diff">
      <div className="diff-hd">
        <span className="tone-accent">{file}</span>
        <span className="tone-ok">+{adds}</span>
        <span className="tone-err">−{dels}</span>
      </div>
      {lines.map((l, i) =>
        l.t === 'hunk' ? (
          <div className="diff-line hunk" key={i}>{l.text}</div>
        ) : (
          <div className={`diff-line ${l.t === 'ctx' ? '' : l.t}`} key={i}>
            <span className="diff-no">{l.a || ''}</span>
            <span className="diff-no">{l.b || ''}</span>
            <span className="diff-sign">{l.t === 'add' ? '+' : l.t === 'del' ? '−' : ' '}</span>
            <span className="diff-text">{l.text}</span>
          </div>
        )
      )}
    </div>
  );
}

Object.assign(window, {
  TagInput, Progress, Spinner, DefList, ConfirmInput, Breadcrumb,
  Tip, PanelState, Pagination, SplitPane, DiffViewer,
});
