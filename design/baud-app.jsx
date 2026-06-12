// baud/ui — demo app: "fleetctl" ops console
const { useState: useStateA, useMemo: useMemoA } = React;

const FLEET = [
  { id: 's1', svc: 'auth-svc', ver: 'v3.2.1', region: 'use1', pods: '9/9', cpu: 82.4, mem: 61.0, rps: 12480, p99: 41, err: 0.02, status: 'ok' },
  { id: 's2', svc: 'billing-api', ver: 'v1.9.0', region: 'use1', pods: '4/4', cpu: 34.1, mem: 44.2, rps: 3211, p99: 87, err: 0.00, status: 'ok' },
  { id: 's3', svc: 'ingest-gw', ver: 'v2.14.0', region: 'euw1', pods: '9/12', cpu: 96.7, mem: 88.9, rps: 48022, p99: 312, err: 4.81, status: 'err' },
  { id: 's4', svc: 'search-idx', ver: 'v5.0.3', region: 'usw2', pods: '6/6', cpu: 58.3, mem: 71.5, rps: 8904, p99: 64, err: 0.11, status: 'warn' },
  { id: 's5', svc: 'search-query', ver: 'v5.0.3', region: 'usw2', pods: '8/8', cpu: 41.7, mem: 52.8, rps: 22310, p99: 38, err: 0.01, status: 'ok' },
  { id: 's6', svc: 'notif-fan', ver: 'v0.8.2', region: 'euw1', pods: '3/3', cpu: 12.9, mem: 23.1, rps: 1502, p99: 18, err: 0.00, status: 'ok' },
  { id: 's7', svc: 'media-proc', ver: 'v4.1.0', region: 'aps1', pods: '5/5', cpu: 77.0, mem: 90.3, rps: 422, p99: 1240, err: 0.92, status: 'warn' },
  { id: 's8', svc: 'edge-cache', ver: 'v7.3.3', region: 'use1', pods: '16/16', cpu: 22.5, mem: 95.1, rps: 91244, p99: 9, err: 0.00, status: 'ok' },
  { id: 's9', svc: 'pg-primary', ver: '15.6', region: 'use1', pods: '1/1', cpu: 61.2, mem: 78.0, rps: 5811, p99: 22, err: 0.00, status: 'ok' },
  { id: 's10', svc: 'kafka-brk', ver: '3.7.0', region: 'use1', pods: '5/5', cpu: 48.9, mem: 66.4, rps: 30122, p99: 14, err: 0.03, status: 'ok' },
  { id: 's11', svc: 'webhook-out', ver: 'v1.2.7', region: 'euw1', pods: '2/2', cpu: 8.3, mem: 19.9, rps: 311, p99: 92, err: 1.20, status: 'warn' },
  { id: 's12', svc: 'flag-svc', ver: 'v2.0.0', region: 'usw2', pods: '3/3', cpu: 5.1, mem: 12.2, rps: 7240, p99: 6, err: 0.00, status: 'ok' },
];

const APP_TREE = [
  { id: 'prod', label: 'prod/', meta: '12 svc', children: [
    { id: 'p-core', label: 'core/', children: [
      { id: 'p-auth', label: 'auth-svc', tone: 'ok', meta: '9' },
      { id: 'p-billing', label: 'billing-api', tone: 'ok', meta: '4' },
      { id: 'p-ingest', label: 'ingest-gw', tone: 'err', meta: '9/12' },
    ]},
    { id: 'p-search', label: 'search/', children: [
      { id: 'p-idx', label: 'search-idx', tone: 'warn', meta: '6' },
      { id: 'p-query', label: 'search-query', tone: 'ok', meta: '8' },
    ]},
    { id: 'p-data', label: 'data/', children: [
      { id: 'p-pg', label: 'pg-primary', tone: 'ok', meta: '1' },
      { id: 'p-kafka', label: 'kafka-brk', tone: 'ok', meta: '5' },
    ]},
  ]},
  { id: 'staging', label: 'staging/', meta: '8 svc', children: [
    { id: 'st-all', label: 'all-services', meta: '8' },
  ]},
];

const LOG_LINES = [
  ['14:32:41.802', 'err', 'ingest-gw-7f9c', 'OOMKilled — container exceeded 2Gi limit'],
  ['14:32:38.114', 'warn', 'ingest-gw-7f9c', 'GC pause 412ms exceeded threshold'],
  ['14:32:35.660', 'info', 'edge-cache-c2a1', 'purged 14,203 keys (deploy hook)'],
  ['14:32:31.090', 'err', 'ingest-gw-44d0', 'liveness probe failed: Get /healthz: context deadline'],
  ['14:32:28.554', 'info', 'auth-svc-9b21', 'rotated signing keys ok'],
  ['14:32:25.013', 'warn', 'media-proc-1e8f', 'transcode queue depth 1,422 (slo 500)'],
  ['14:32:21.371', 'info', 'kafka-brk-3', 'ISR shrink topic=events partition=7'],
  ['14:32:18.876', 'info', 'flag-svc-0a44', 'flag "new-ranker" 25% → 50%'],
  ['14:32:14.209', 'err', 'ingest-gw-44d0', 'upstream connect error: 503 UC upstream_reset',],
  ['14:32:09.991', 'info', 'webhook-out-77b2', 'retry budget exhausted for dest=hooks.stripe.com'],
];

const TONE_BY_STATUS = { ok: 'ok', warn: 'warn', err: 'err' };

function MetricCell({ label, value, sub, tone }) {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 1, padding: 'var(--pad)', borderRight: '1px solid var(--border)', minWidth: 0 }}>
      <span className="field-label">{label}</span>
      <span className={tone ? `tone-${tone}` : ''} style={{ fontSize: 'calc(var(--fs) * 1.5)', fontWeight: 700, lineHeight: 1.15, fontVariantNumeric: 'tabular-nums' }}>{value}</span>
      <span className="tone-faint" style={{ fontSize: 'var(--fs-sm)' }}>{sub}</span>
    </div>
  );
}

function AppView({ push, openPalette }) {
  const [tab, setTab] = useStateA('fleet');
  const [filter, setFilter] = useStateA('');
  const [region, setRegion] = useStateA('all');
  const [drawer, setDrawer] = useStateA(null);
  const [killModal, setKillModal] = useStateA(false);

  const rows = useMemoA(() => FLEET.filter((r) =>
    (region === 'all' || r.region === region) &&
    (!filter || r.svc.includes(filter.toLowerCase()))
  ), [filter, region]);

  const counts = useMemoA(() => ({
    ok: FLEET.filter((r) => r.status === 'ok').length,
    warn: FLEET.filter((r) => r.status === 'warn').length,
    err: FLEET.filter((r) => r.status === 'err').length,
  }), []);

  return (
    <div className="app-shell" data-screen-label="Demo app: fleetctl">
      {/* top bar */}
      <div className="app-top">
        <span className="brand">
          <span className="brand-name">fleetctl</span>
          <span className="brand-sub">prod · us-east-1 primary</span>
        </span>
        <Tabs value={tab} onChange={setTab} items={[
          { id: 'fleet', label: 'Fleet', badge: 12 },
          { id: 'incidents', label: 'Incidents', badge: 1 },
          { id: 'deploys', label: 'Deploys' },
        ]} />
        <span style={{ flex: 1 }}></span>
        <Input placeholder="filter services…  /" value={filter} onChange={(e) => setFilter(e.target.value)} style={{ width: 190 }} />
        <Select value={region} onChange={setRegion} width={104} alignRight options={[
          { value: 'all', label: 'all regions' },
          { value: 'use1', label: 'us-east-1' },
          { value: 'usw2', label: 'us-west-2' },
          { value: 'euw1', label: 'eu-west-1' },
          { value: 'aps1', label: 'ap-south-1' },
        ]} />
        <Btn variant="primary" kbd="⌘K" onClick={openPalette}>Cmd</Btn>
      </div>

      {/* metrics strip */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(6, 1fr)', borderBottom: '1px solid var(--border)', background: 'var(--bg-panel)' }}>
        <MetricCell label="Fleet RPS" value="234,179" sub="+4.2% vs 1h ago" />
        <MetricCell label="p99 latency" value="312 ms" sub="SLO 250ms" tone="err" />
        <MetricCell label="Error rate" value="0.61%" sub="budget 38% burned" tone="warn" />
        <MetricCell label="Pods" value="71 / 74" sub="3 not ready" tone="warn" />
        <MetricCell label="Healthy" value={`${counts.ok}/12`} sub={`${counts.warn} warn · ${counts.err} down`} tone="ok" />
        <MetricCell label="Active incident" value="SEV-1" sub="INC-2291 · 14m" tone="err" />
      </div>

      {/* main split */}
      <div style={{ flex: 1, minHeight: 0, display: 'grid', gridTemplateColumns: '220px 1fr', gap: 0 }}>
        <div className="panel" style={{ borderTop: 'none', borderLeft: 'none', borderBottom: 'none' }}>
          <div className="panel-hd"><span className="panel-title">Navigator</span><span className="panel-acts"><Kbd>j/k</Kbd></span></div>
          <div className="panel-bd" style={{ padding: '3px 0' }}>
            <Tree nodes={APP_TREE} defaultOpen={['prod', 'p-core']} defaultSel="p-ingest" />
          </div>
        </div>

        <div style={{ display: 'flex', flexDirection: 'column', minWidth: 0 }}>
          {/* fleet table */}
          <div className="panel" style={{ flex: 1.6, minHeight: 0, border: 'none', borderBottom: '1px solid var(--border)' }}>
            <div className="panel-hd">
              <Breadcrumb onNav={(it) => push('info', `Navigate to ${it.label}`)} items={[
                { id: 'prod', label: 'prod' },
                { id: 'core', label: 'core' },
                { id: 'svc', label: 'services' },
              ]} />
              <span className="panel-acts">
                <span className="panel-title" style={{ marginRight: 6 }}>{rows.length} shown</span>
                <Btn variant="ghost" onClick={() => push('info', 'Snapshot exported to s3://ops-dumps/')}>Export</Btn>
                <Btn variant="danger" onClick={() => setKillModal(true)}>Kill pod</Btn>
              </span>
            </div>
            <div className="panel-bd">
              <DataTable
                zebra
                initialSort={{ key: 'err', dir: 'desc' }}
                columns={[
                  { key: 'svc', label: 'Service', render: (r) => <span style={{ cursor: 'pointer' }} onClick={(e) => { e.stopPropagation(); setDrawer(r); }}>{r.svc}</span> },
                  { key: 'ver', label: 'Ver' },
                  { key: 'region', label: 'Reg' },
                  { key: 'pods', label: 'Pods', num: true },
                  { key: 'cpu', label: 'CPU%', num: true, render: (r) => <span className={r.cpu > 90 ? 'tone-err' : r.cpu > 70 ? 'tone-warn' : ''}>{r.cpu.toFixed(1)}</span> },
                  { key: 'mem', label: 'Mem%', num: true, render: (r) => <span className={r.mem > 90 ? 'tone-err' : r.mem > 75 ? 'tone-warn' : ''}>{r.mem.toFixed(1)}</span> },
                  { key: 'rps', label: 'RPS', num: true, render: (r) => r.rps.toLocaleString() },
                  { key: 'p99', label: 'p99', num: true, render: (r) => <span className={r.p99 > 300 ? 'tone-err' : r.p99 > 100 ? 'tone-warn' : ''}>{r.p99}</span> },
                  { key: 'err', label: 'Err%', num: true, render: (r) => <span className={r.err > 1 ? 'tone-err' : r.err > 0.1 ? 'tone-warn' : 'tone-faint'}>{r.err.toFixed(2)}</span> },
                  { key: 'status', label: 'St', render: (r) => <Badge tone={TONE_BY_STATUS[r.status]} dot>{r.status}</Badge> },
                ]}
                rows={rows}
              />
            </div>
            <Pagination page={1} pageSize={12} total={74} onPage={() => {}}
                        onMore={() => push('info', 'Fetched next page (htmx hx-get append)')} />
          </div>

          {/* log tail */}
          <div className="panel" style={{ flex: 1, minHeight: 0, border: 'none' }}>
            <div className="panel-hd">
              <span className="panel-title">Log tail — ingest-gw</span>
              <span className="panel-acts">
                <Badge tone="err" variant="outline">err only</Badge>
                <Dot tone="ok" pulse />
                <span className="tone-faint" style={{ fontSize: 'var(--fs-sm)' }}>streaming</span>
              </span>
            </div>
            <div className="panel-bd" style={{ padding: '2px 0', fontVariantNumeric: 'tabular-nums' }}>
              {LOG_LINES.map((l, i) => (
                <div key={i} style={{ display: 'flex', gap: 10, padding: '0 var(--pad)', height: 'var(--row)', alignItems: 'center', whiteSpace: 'nowrap', overflow: 'hidden' }}>
                  <span className="tone-faint">{l[0]}</span>
                  <span className={`tone-${l[1]}`} style={{ width: 38, flex: 'none', textTransform: 'uppercase', fontSize: 'var(--fs-sm)', fontWeight: 700 }}>{l[1]}</span>
                  <span className="tone-accent" style={{ width: 140, flex: 'none', overflow: 'hidden', textOverflow: 'ellipsis' }}>{l[2]}</span>
                  <span style={{ overflow: 'hidden', textOverflow: 'ellipsis' }}>{l[3]}</span>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>

      {drawer ? (
        <Drawer title={`${drawer.svc} · ${drawer.ver}`} onClose={() => setDrawer(null)}>
          <div className="sheet-row">
            <Badge tone={TONE_BY_STATUS[drawer.status]} dot>{drawer.status}</Badge>
            <span className="tone-faint">{drawer.pods} pods · {drawer.region}</span>
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 8 }}>
            <Field label="CPU"><span className={drawer.cpu > 90 ? 'tone-err' : ''}>{drawer.cpu.toFixed(1)}%</span></Field>
            <Field label="Memory"><span className={drawer.mem > 90 ? 'tone-err' : ''}>{drawer.mem.toFixed(1)}%</span></Field>
            <Field label="RPS"><span>{drawer.rps.toLocaleString()}</span></Field>
            <Field label="p99"><span className={drawer.p99 > 300 ? 'tone-err' : ''}>{drawer.p99} ms</span></Field>
          </div>
          <div className="sheet-row">
            <Btn variant="primary" onClick={() => { setDrawer(null); push('ok', `Restart queued: ${drawer.svc}`); }}>Restart</Btn>
            <Btn onClick={() => push('info', `Tailing logs for ${drawer.svc}`)}>Logs</Btn>
            <Btn variant="danger" onClick={() => { setDrawer(null); setKillModal(true); }}>Kill pod</Btn>
          </div>
        </Drawer>
      ) : null}

      {killModal ? (
        <Modal title="Kill pod" onClose={() => setKillModal(false)} footer={
          <React.Fragment>
            <Btn variant="ghost" onClick={() => setKillModal(false)}>Cancel <Kbd>esc</Kbd></Btn>
            <Btn variant="danger" onClick={() => { setKillModal(false); push('ok', 'Pod ingest-gw-7f9c terminated'); }}>Kill ingest-gw-7f9c</Btn>
          </React.Fragment>
        }>
          <p>Terminate pod <span className="tone-err">ingest-gw-7f9c</span>? The replicaset will reschedule it.</p>
          <p className="tone-faint" style={{ fontSize: 'var(--fs-sm)' }}>Pod is OOMKilled-looping · 6 restarts in 14m · node use1-c4.2xl-09</p>
        </Modal>
      ) : null}
    </div>
  );
}

window.AppView = AppView;
