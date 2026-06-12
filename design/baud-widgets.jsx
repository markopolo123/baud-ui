// baud/ui — widgets: DataTable, Tree, StatusBar, Modal, Drawer, Palette, Toasts
const { useState: useStateW, useEffect: useEffectW, useRef: useRefW, useMemo: useMemoW } = React;

/* ---------------- DataTable ---------------- */
// columns: [{key, label, num, render?, width?}]
function DataTable({ columns, rows, zebra, lines, selectable = true, initialSort, rowKey = 'id' }) {
  const [sort, setSort] = useStateW(initialSort || null); // {key, dir}
  const [selected, setSelected] = useStateW(null);
  const sorted = useMemoW(() => {
    if (!sort) return rows;
    const r = [...rows].sort((a, b) => {
      const av = a[sort.key], bv = b[sort.key];
      if (typeof av === 'number' && typeof bv === 'number') return av - bv;
      return String(av).localeCompare(String(bv));
    });
    return sort.dir === 'desc' ? r.reverse() : r;
  }, [rows, sort]);
  const clickSort = (key) => {
    setSort((s) => (!s || s.key !== key) ? { key, dir: 'asc' } : s.dir === 'asc' ? { key, dir: 'desc' } : null);
  };
  const cls = ['dt'];
  if (zebra) cls.push('dt-zebra');
  if (lines) cls.push('dt-lines');
  return (
    <table className={cls.join(' ')}>
      <thead>
        <tr>
          <th className="row-mark" style={{ cursor: 'default' }}></th>
          {columns.map((c) => (
            <th key={c.key} className={`${c.num ? 'num ' : ''}${sort && sort.key === c.key ? 'sorted' : ''}`}
                style={c.width ? { width: c.width } : null}
                onClick={() => clickSort(c.key)}>
              {c.label}
              {sort && sort.key === c.key ? <span className="sort-arrow">{sort.dir === 'asc' ? '▲' : '▼'}</span> : null}
            </th>
          ))}
        </tr>
      </thead>
      <tbody>
        {sorted.map((r) => (
          <tr key={r[rowKey]} className={selected === r[rowKey] ? 'sel' : ''}
              onClick={selectable ? () => setSelected(r[rowKey]) : null}>
            <td className="row-mark">{selected === r[rowKey] ? '▌' : ' '}</td>
            {columns.map((c) => (
              <td key={c.key} className={c.num ? 'num' : ''}>
                {c.render ? c.render(r) : r[c.key]}
              </td>
            ))}
          </tr>
        ))}
      </tbody>
    </table>
  );
}

/* ---------------- Tree ---------------- */
// nodes: [{id, label, meta?, tone?, children?}]
function Tree({ nodes, defaultOpen = [], defaultSel }) {
  const [open, setOpen] = useStateW(() => new Set(defaultOpen));
  const [sel, setSel] = useStateW(defaultSel || null);
  const toggle = (id) => setOpen((s) => { const n = new Set(s); n.has(id) ? n.delete(id) : n.add(id); return n; });
  const renderNodes = (list, depth, prefix) =>
    list.map((n, i) => {
      const last = i === list.length - 1;
      const branch = depth === 0 ? '' : prefix + (last ? '└─' : '├─');
      const childPrefix = depth === 0 ? '' : prefix + (last ? '   ' : '│  ');
      const isDir = !!n.children;
      const isOpen = open.has(n.id);
      return (
        <React.Fragment key={n.id}>
          <li className={`tree-row${sel === n.id ? ' sel' : ''}`}
              onClick={() => { setSel(n.id); if (isDir) toggle(n.id); }}>
            <span className="tree-glyph">{branch}{isDir ? (isOpen ? '▾ ' : '▸ ') : '  '}</span>
            <span className={n.tone ? `tone-${n.tone}` : ''}>{n.label}</span>
            {n.meta ? <span className="tree-meta">{n.meta}</span> : null}
          </li>
          {isDir && isOpen ? renderNodes(n.children, depth + 1, childPrefix) : null}
        </React.Fragment>
      );
    });
  return <ul className="tree">{renderNodes(nodes, 0, '')}</ul>;
}

/* ---------------- StatusBar ---------------- */
// cells: [{id, content, mode?, spring?}]
function StatusBar({ cells }) {
  return (
    <div className="statusbar">
      {cells.map((c) =>
        c.spring ? <span key={c.id} className="sb-cell sb-spring"></span> : (
          <span key={c.id} className={`sb-cell${c.mode ? ' sb-mode' : ''}`}>{c.content}</span>
        )
      )}
    </div>
  );
}

/* ---------------- Modal ---------------- */
function Modal({ title, onClose, footer, children }) {
  useEffectW(() => {
    const h = (e) => { if (e.key === 'Escape') onClose(); };
    document.addEventListener('keydown', h);
    return () => document.removeEventListener('keydown', h);
  }, [onClose]);
  return (
    <div className="overlay" onMouseDown={(e) => { if (e.target === e.currentTarget) onClose(); }}>
      <div className="modal">
        <div className="modal-hd">
          <span className="modal-title">{title}</span>
          <button className="x-btn" onClick={onClose}>✕</button>
        </div>
        <div className="modal-bd">{children}</div>
        {footer ? <div className="modal-ft">{footer}</div> : null}
      </div>
    </div>
  );
}

/* ---------------- Drawer ---------------- */
function Drawer({ title, onClose, children }) {
  useEffectW(() => {
    const h = (e) => { if (e.key === 'Escape') onClose(); };
    document.addEventListener('keydown', h);
    return () => document.removeEventListener('keydown', h);
  }, [onClose]);
  return (
    <div className="drawer-wrap" onMouseDown={(e) => { if (e.target === e.currentTarget) onClose(); }}>
      <div className="drawer">
        <div className="modal-hd">
          <span className="modal-title">{title}</span>
          <button className="x-btn" onClick={onClose}>✕</button>
        </div>
        <div className="panel-bd" style={{ padding: 10, display: 'flex', flexDirection: 'column', gap: 10 }}>
          {children}
        </div>
      </div>
    </div>
  );
}

/* ---------------- Command Palette ---------------- */
// commands: [{id, cat, label, keys?, run?}]
function Palette({ commands, onClose, onRun }) {
  const [q, setQ] = useStateW('');
  const [hl, setHl] = useStateW(0);
  const inputRef = useRefW(null);
  useEffectW(() => { inputRef.current && inputRef.current.focus(); }, []);
  const filtered = useMemoW(() => {
    const t = q.trim().toLowerCase();
    if (!t) return commands;
    return commands.filter((c) => (c.cat + ' ' + c.label).toLowerCase().includes(t));
  }, [q, commands]);
  useEffectW(() => { setHl(0); }, [q]);
  const key = (e) => {
    if (e.key === 'Escape') onClose();
    else if (e.key === 'ArrowDown') { e.preventDefault(); setHl((h) => Math.min(h + 1, filtered.length - 1)); }
    else if (e.key === 'ArrowUp') { e.preventDefault(); setHl((h) => Math.max(h - 1, 0)); }
    else if (e.key === 'Enter' && filtered[hl]) { onRun(filtered[hl]); }
  };
  return (
    <div className="overlay" onMouseDown={(e) => { if (e.target === e.currentTarget) onClose(); }}>
      <div className="palette">
        <div className="palette-input">
          <span className="prompt">›</span>
          <input ref={inputRef} className="input" placeholder="Type a command… (esc to close)"
                 value={q} onChange={(e) => setQ(e.target.value)} onKeyDown={key} />
          <Kbd>↑↓</Kbd><Kbd>↵</Kbd>
        </div>
        <div className="palette-list">
          {filtered.length === 0 ? <div className="palette-empty">No commands match “{q}”</div> : null}
          {filtered.map((c, i) => (
            <button key={c.id} className={`palette-item${i === hl ? ' hl' : ''}`}
                    onMouseEnter={() => setHl(i)} onClick={() => onRun(c)}>
              <span className="pi-cat">{c.cat}</span>
              <span>{c.label}</span>
              {c.keys ? <span className="pi-keys">{c.keys.map((k) => <Kbd key={k}>{k}</Kbd>)}</span> : null}
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}

/* ---------------- Toasts ---------------- */
const TOAST_GLYPH = { ok: '✓', err: '✗', warn: '▲', info: 'ℹ' };
function Toasts({ items, onDismiss }) {
  return (
    <div className="toasts">
      {items.map((t) => (
        <div key={t.id} className={`toast tone-${t.tone}`}>
          <span className="toast-glyph">{TOAST_GLYPH[t.tone] || '·'}</span>
          <span className="toast-msg">{t.msg}</span>
          <button className="x-btn" onClick={() => onDismiss(t.id)}>✕</button>
        </div>
      ))}
    </div>
  );
}

function useToasts() {
  const [items, setItems] = useStateW([]);
  const idRef = useRefW(0);
  const push = (tone, msg) => {
    const id = ++idRef.current;
    setItems((t) => [...t, { id, tone, msg }]);
    setTimeout(() => setItems((t) => t.filter((x) => x.id !== id)), 4200);
  };
  const dismiss = (id) => setItems((t) => t.filter((x) => x.id !== id));
  return { toasts: items, push, dismiss };
}

Object.assign(window, { DataTable, Tree, StatusBar, Modal, Drawer, Palette, Toasts, useToasts });
