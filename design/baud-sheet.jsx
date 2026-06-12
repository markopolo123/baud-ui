// baud/ui — component sheet view (kitchen sink)
const { useState: useStateS } = React;

const SHEET_SWATCHES = [
  ['bg-app', '--bg-app'], ['bg-panel', '--bg-panel'], ['bg-raised', '--bg-raised'],
  ['border', '--border'], ['fg', '--fg'], ['fg-muted', '--fg-muted'],
  ['accent', '--accent'], ['accent-2', '--accent-2'],
  ['ok', '--ok'], ['warn', '--warn'], ['err', '--err'], ['info', '--info'],
];

const SHEET_ROWS = [
  { id: 'r1', sym: 'auth-svc', region: 'us-east-1', cpu: 82.4, mem: 61.0, rps: 12480, p99: 41, status: 'ok' },
  { id: 'r2', sym: 'billing-api', region: 'us-east-1', cpu: 34.1, mem: 44.2, rps: 3211, p99: 87, status: 'ok' },
  { id: 'r3', sym: 'ingest-gw', region: 'eu-west-1', cpu: 96.7, mem: 88.9, rps: 48022, p99: 312, status: 'err' },
  { id: 'r4', sym: 'search-idx', region: 'us-west-2', cpu: 58.3, mem: 71.5, rps: 8904, p99: 64, status: 'warn' },
  { id: 'r5', sym: 'notif-fan', region: 'eu-west-1', cpu: 12.9, mem: 23.1, rps: 1502, p99: 18, status: 'ok' },
  { id: 'r6', sym: 'media-proc', region: 'ap-south-1', cpu: 77.0, mem: 90.3, rps: 422, p99: 1240, status: 'warn' },
];

const STATUS_TONE = { ok: 'ok', warn: 'warn', err: 'err' };

const SHEET_TREE = [
  { id: 'svc', label: 'services/', meta: '14', children: [
    { id: 'svc-auth', label: 'auth-svc', meta: '9 pods', tone: 'ok' },
    { id: 'svc-billing', label: 'billing-api', meta: '4 pods', tone: 'ok' },
    { id: 'svc-ingest', label: 'ingest-gw', meta: '12 pods', tone: 'err' },
    { id: 'svc-search', label: 'search/', children: [
      { id: 'svc-search-idx', label: 'search-idx', meta: '6 pods', tone: 'warn' },
      { id: 'svc-search-q', label: 'search-query', meta: '8 pods', tone: 'ok' },
    ]},
  ]},
  { id: 'jobs', label: 'jobs/', meta: '3', children: [
    { id: 'job-etl', label: 'etl-nightly', meta: 'cron' },
    { id: 'job-bk', label: 'backup-pg', meta: 'cron' },
  ]},
  { id: 'cfg', label: 'config/', meta: '2', children: [
    { id: 'cfg-flags', label: 'feature-flags.toml' },
    { id: 'cfg-limits', label: 'rate-limits.toml' },
  ]},
];

function SheetView({ push }) {
  const [tab1, setTab1] = useStateS('pods');
  const [tab2, setTab2] = useStateS('1h');
  const [env, setEnv] = useStateS('prod');
  const [owner, setOwner] = useStateS('team-platform');
  const [since, setSince] = useStateS(() => new Date());
  const [proto, setProto] = useStateS('grpc');
  const [chk1, setChk1] = useStateS(true);
  const [chk2, setChk2] = useStateS(false);
  const [radio, setRadio] = useStateS('rolling');
  const [modal, setModal] = useStateS(false);
  const [drawer, setDrawer] = useStateS(false);
  const [pop, setPop] = useStateS(false);
  const popRef = React.useRef(null);
  useClickOutside(popRef, () => setPop(false), pop);

  return (
    <div style={{ padding: 14, display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16, alignContent: 'start' }}>

      {/* colors */}
      <section className="sheet-section" style={{ gridColumn: '1 / -1' }} data-screen-label="Sheet: Colors">
        <h2 className="sheet-h">Theme tokens</h2>
        <div className="sheet-row">
          {SHEET_SWATCHES.map(([name, v]) => (
            <div className="swatch" key={name}>
              <div className="swatch-chip" style={{ background: `var(${v})` }}></div>
              <span className="swatch-name">{name}</span>
            </div>
          ))}
        </div>
      </section>

      {/* buttons */}
      <section className="sheet-section" data-screen-label="Sheet: Buttons">
        <h2 className="sheet-h">Buttons · toolbar</h2>
        <div className="sheet-row">
          <Btn variant="primary" kbd="⏎">Deploy</Btn>
          <Btn>Restart</Btn>
          <Btn glyph="◌">Scale</Btn>
          <Btn variant="danger">Kill</Btn>
          <Btn variant="ghost">Logs</Btn>
          <Btn disabled>Rollback</Btn>
        </div>
        <div className="sheet-row">
          <BtnGroup>
            <Btn active>1m</Btn><Btn>5m</Btn><Btn>1h</Btn><Btn>1d</Btn>
          </BtnGroup>
          <span className="pop-wrap" ref={popRef}>
            <Btn glyph="▾" onClick={() => setPop(!pop)}>Actions</Btn>
            {pop ? (
              <div className="popover">
                <span className="field-label">Quick actions</span>
                <button className="menu-item" onClick={() => { setPop(false); push('ok', 'Cache flushed on 12 pods'); }}><span>Flush cache</span><Kbd>f</Kbd></button>
                <button className="menu-item" onClick={() => { setPop(false); push('info', 'Draining traffic from eu-west-1'); }}><span>Drain traffic</span><Kbd>d</Kbd></button>
                <button className="menu-item" onClick={() => { setPop(false); push('warn', 'Maintenance mode armed'); }}><span>Maintenance mode</span></button>
              </div>
            ) : null}
          </span>
        </div>
      </section>

      {/* badges */}
      <section className="sheet-section" data-screen-label="Sheet: Badges">
        <h2 className="sheet-h">Badges · indicators</h2>
        <div className="sheet-row">
          <Badge tone="ok" dot>Healthy</Badge>
          <Badge tone="warn" dot>Degraded</Badge>
          <Badge tone="err" dot>Down</Badge>
          <Badge tone="info">Syncing</Badge>
          <Badge tone="accent">Canary</Badge>
          <Badge tone="neutral">Idle</Badge>
        </div>
        <div className="sheet-row">
          <Badge tone="ok" variant="solid">200</Badge>
          <Badge tone="warn" variant="solid">429</Badge>
          <Badge tone="err" variant="solid">503</Badge>
          <Badge tone="err" variant="outline">SEV-1</Badge>
          <Badge tone="warn" variant="outline">SEV-2</Badge>
          <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}>
            <Dot tone="ok" pulse /> <span className="tone-faint">live</span>
          </span>
        </div>
      </section>

      {/* forms */}
      <section className="sheet-section" data-screen-label="Sheet: Forms">
        <h2 className="sheet-h">Forms · inputs</h2>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 10 }}>
          <Field label="Service name" hint="lowercase, dashes only">
            <Input placeholder="my-service" defaultValue="ingest-gw" />
          </Field>
          <Field label="Replicas" error="exceeds quota (max 24)">
            <Input defaultValue="32" error suffix="pods" />
          </Field>
          <Field label="Environment">
            <Select value={env} onChange={setEnv} options={[
              { value: 'prod', label: 'production' },
              { value: 'staging', label: 'staging' },
              { value: 'dev', label: 'dev' },
            ]} />
          </Field>
          <Field label="Endpoint" hint="health check target">
            <Input prefix="https://" defaultValue="api.internal" suffix="/healthz" />
          </Field>
          <Field label="Owner team" hint="combobox — type to filter">
            <Combobox value={owner} onChange={setOwner} placeholder="search teams…" options={[
              { value: 'team-platform', label: 'team-platform', meta: '14 svc' },
              { value: 'team-payments', label: 'team-payments', meta: '6 svc' },
              { value: 'team-search', label: 'team-search', meta: '5 svc' },
              { value: 'team-media', label: 'team-media', meta: '3 svc' },
              { value: 'team-growth', label: 'team-growth', meta: '2 svc' },
              { value: 'team-sre', label: 'team-sre', meta: 'on-call' },
            ]} />
          </Field>
          <Field label="Logs since" hint="date picker — presets below grid">
            <DatePicker value={since} onChange={setSince} />
          </Field>
        </div>
        <div className="sheet-row" style={{ marginTop: 2 }}>
          <Checkbox checked={chk1} onChange={setChk1}>auto-restart</Checkbox>
          <Checkbox checked={chk2} onChange={setChk2}>verbose logs</Checkbox>
          <Checkbox checked disabled onChange={() => {}}>mTLS (enforced)</Checkbox>
        </div>
        <div className="sheet-row">
          <Radio checked={radio === 'rolling'} onChange={() => setRadio('rolling')}>rolling</Radio>
          <Radio checked={radio === 'bluegreen'} onChange={() => setRadio('bluegreen')}>blue/green</Radio>
          <Radio checked={radio === 'recreate'} onChange={() => setRadio('recreate')}>recreate</Radio>
          <Toggle options={['grpc', 'http', 'tcp']} value={proto} onChange={setProto} />
        </div>
      </section>

      {/* tabs */}
      <section className="sheet-section" data-screen-label="Sheet: Tabs">
        <h2 className="sheet-h">Tabs</h2>
        <Tabs value={tab1} onChange={setTab1} items={[
          { id: 'pods', label: 'Pods', badge: 39 },
          { id: 'events', label: 'Events', badge: 7 },
          { id: 'logs', label: 'Logs' },
          { id: 'env', label: 'Env' },
        ]} />
        <Tabs boxed value={tab2} onChange={setTab2} items={[
          { id: '5m', label: '5m' }, { id: '1h', label: '1h' },
          { id: '1d', label: '1d' }, { id: '1w', label: '1w' },
        ]} />
      </section>

      {/* overlays */}
      <section className="sheet-section" data-screen-label="Sheet: Overlays">
        <h2 className="sheet-h">Overlays · feedback</h2>
        <div className="sheet-row">
          <Btn onClick={() => setModal(true)}>Open modal</Btn>
          <Btn onClick={() => setDrawer(true)}>Open drawer</Btn>
          <Btn onClick={() => push('ok', 'Deployed ingest-gw v2.14.1 to prod')}>Toast ✓</Btn>
          <Btn onClick={() => push('err', 'Connection lost: eu-west-1 etcd quorum')}>Toast ✗</Btn>
          <Btn onClick={() => push('warn', 'p99 latency above SLO for 5m')}>Toast ▲</Btn>
        </div>
        <p className="tone-faint" style={{ fontSize: 'var(--fs-sm)' }}>
          Command palette: <Kbd>⌘K</Kbd> anywhere · modals close on <Kbd>esc</Kbd>
        </p>
      </section>

      {/* table */}
      <section className="sheet-section" data-screen-label="Sheet: Table">
        <h2 className="sheet-h">Data table (click headers to sort)</h2>
        <div className="panel">
          <div className="panel-bd" style={{ maxHeight: 240 }}>
            <DataTable
              zebra lines
              initialSort={{ key: 'cpu', dir: 'desc' }}
              columns={[
                { key: 'sym', label: 'Service' },
                { key: 'region', label: 'Region' },
                { key: 'cpu', label: 'CPU%', num: true, render: (r) => <span className={r.cpu > 90 ? 'tone-err' : r.cpu > 70 ? 'tone-warn' : ''}>{r.cpu.toFixed(1)}</span> },
                { key: 'mem', label: 'Mem%', num: true, render: (r) => r.mem.toFixed(1) },
                { key: 'rps', label: 'RPS', num: true, render: (r) => r.rps.toLocaleString() },
                { key: 'p99', label: 'p99 ms', num: true, render: (r) => <span className={r.p99 > 300 ? 'tone-err' : ''}>{r.p99}</span> },
                { key: 'status', label: 'St', render: (r) => <Badge tone={STATUS_TONE[r.status]} dot>{r.status}</Badge> },
              ]}
              rows={SHEET_ROWS}
            />
          </div>
        </div>
      </section>

      {/* tree */}
      <section className="sheet-section" data-screen-label="Sheet: Tree">
        <h2 className="sheet-h">Tree view</h2>
        <div className="panel">
          <div className="panel-bd" style={{ maxHeight: 240, padding: '3px 0' }}>
            <Tree nodes={SHEET_TREE} defaultOpen={['svc', 'svc-search']} defaultSel="svc-ingest" />
          </div>
        </div>
      </section>

      {/* extras: tags, progress, breadcrumb, tooltip, pagination, deflist,
          confirm, states, splitpane, diff */}
      <SheetExtras push={push} />

      {modal ? (
        <Modal title="Confirm deploy" onClose={() => setModal(false)} footer={
          <React.Fragment>
            <Btn variant="ghost" onClick={() => setModal(false)}>Cancel <Kbd>esc</Kbd></Btn>
            <Btn variant="primary" kbd="⏎" onClick={() => { setModal(false); push('ok', 'Deploy queued: ingest-gw v2.14.1'); }}>Deploy</Btn>
          </React.Fragment>
        }>
          <p>Deploy <span className="tone-accent">ingest-gw v2.14.1</span> to <span className="tone-warn">production</span>?</p>
          <Field label="Strategy"><Toggle options={['rolling', 'blue/green']} value="rolling" onChange={() => {}} /></Field>
          <p className="tone-faint" style={{ fontSize: 'var(--fs-sm)' }}>12 pods will be replaced · est. 4m 30s · last deploy 2d ago by mira@</p>
        </Modal>
      ) : null}

      {drawer ? (
        <Drawer title="ingest-gw · details" onClose={() => setDrawer(false)}>
          <div className="sheet-row">
            <Badge tone="err" dot>Down</Badge>
            <Badge tone="accent">Canary</Badge>
            <span className="tone-faint">v2.14.0 · 12 pods</span>
          </div>
          <Field label="Owner"><Input defaultValue="platform-team" /></Field>
          <Field label="Runbook"><Input prefix="https://" defaultValue="wiki/runbooks/ingest-gw" /></Field>
          <Panel title="Recent events" bodyStyle={{ padding: 6 }}>
            <p className="tone-err" style={{ fontSize: 'var(--fs-sm)' }}>14:32:08 OOMKilled pod ingest-gw-7f9c</p>
            <p className="tone-warn" style={{ fontSize: 'var(--fs-sm)' }}>14:31:55 Liveness probe failed ×3</p>
            <p className="tone-faint" style={{ fontSize: 'var(--fs-sm)' }}>14:28:10 Scaled 10 → 12</p>
          </Panel>
        </Drawer>
      ) : null}
    </div>
  );
}

window.SheetView = SheetView;
