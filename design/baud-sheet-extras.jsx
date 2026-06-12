// baud/ui — sheet sections for the extras components
const { useState: useStateE, useEffect: useEffectE } = React;

const DIFF_SAMPLE = [
  { t: 'hunk', text: '@@ -12,7 +12,9 @@ resources:' },
  { t: 'ctx', a: 12, b: 12, text: '  limits:' },
  { t: 'del', a: 13, text: '    memory: 2Gi' },
  { t: 'add', b: 13, text: '    memory: 4Gi' },
  { t: 'add', b: 14, text: '    ephemeral-storage: 1Gi' },
  { t: 'ctx', a: 14, b: 15, text: '  requests:' },
  { t: 'ctx', a: 15, b: 16, text: '    cpu: 500m' },
  { t: 'del', a: 16, text: '    memory: 1Gi' },
  { t: 'add', b: 17, text: '    memory: 2Gi' },
  { t: 'ctx', a: 17, b: 18, text: 'replicas: 12' },
];

function SheetExtras({ push }) {
  const [tags, setTags] = useStateE(['env=prod', 'region=euw1']);
  const [stateTab, setStateTab] = useStateE('skeleton');
  const [page, setPage] = useStateE(3);
  const [deploying, setDeploying] = useStateE(34);
  useEffectE(() => {
    if (window.matchMedia && window.matchMedia('(prefers-reduced-motion: reduce)').matches) return;
    const t = setInterval(() => setDeploying((d) => (d >= 100 ? 0 : d + 1)), 120);
    return () => clearInterval(t);
  }, []);

  return (
    <React.Fragment>

      {/* tags + progress */}
      <section className="sheet-section" data-screen-label="Sheet: Tags & progress">
        <h2 className="sheet-h">Multi-select tags · progress</h2>
        <Field label="Filter labels" hint="type and ⏎, or pick — backspace deletes">
          <TagInput value={tags} onChange={setTags} placeholder="key=value…" suggestions={[
            'env=prod', 'env=staging', 'region=use1', 'region=usw2', 'region=euw1',
            'tier=critical', 'team=platform', 'team=payments', 'canary=true',
          ]} />
        </Field>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
          <Progress label="rollout ingest-gw" value={deploying} />
          <Progress label="pg-primary disk" value={87} tone="warn" />
          <Progress label="etl-nightly" value={100} tone="ok" />
          <div className="sheet-row" style={{ marginTop: 2 }}>
            <span><Spinner /> <span className="tone-faint">deploying…</span></span>
            <span><Spinner tone="warn" /> <span className="tone-faint">draining…</span></span>
            <span className="tone-faint" style={{ fontSize: 'var(--fs-sm)' }}>braille spinner · ▰▱ ascii bars</span>
          </div>
        </div>
      </section>

      {/* breadcrumb + tooltip + pagination */}
      <section className="sheet-section" data-screen-label="Sheet: Breadcrumb, tooltip, pagination">
        <h2 className="sheet-h">Breadcrumb · tooltip · pagination</h2>
        <Breadcrumb onNav={(it) => push('info', `Navigate to ${it.label}`)} items={[
          { id: 'prod', label: 'prod' },
          { id: 'core', label: 'core' },
          { id: 'ingest', label: 'ingest-gw' },
          { id: 'pod', label: 'ingest-gw-7f9c' },
        ]} />
        <p>
          Latency is <Tip under tip={'p99 = 312ms\np95 = 104ms\np50 =  22ms'}>over SLO</Tip> and the error
          budget is <Tip under tip="38.2% of 30d budget consumed">burning</Tip>.
          Hover a <Tip tip="works on any element"><Badge tone="info">badge</Badge></Tip> too.
        </p>
        <div className="panel">
          <Pagination page={page} pageSize={50} total={12403} onPage={setPage}
                      onMore={() => push('info', 'Fetched 50 more rows (htmx hx-get append)')} />
        </div>
      </section>

      {/* deflist + confirm */}
      <section className="sheet-section" data-screen-label="Sheet: Definition list & confirm">
        <h2 className="sheet-h">Definition list · destructive confirm</h2>
        <div className="panel">
          <div className="panel-bd" style={{ padding: 'var(--pad)' }}>
            <DefList lines rows={[
              { k: 'service', v: 'ingest-gw' },
              { k: 'version', v: 'v2.14.0 → v2.14.1', tone: 'accent' },
              { k: 'uptime', v: '14d 6h 12m' },
              { k: 'restarts', v: '6 in 14m', tone: 'err' },
              { k: 'node', v: 'use1-c4.2xl-09' },
              { k: 'last deploy', v: '2026-06-09 11:40 by mira@' },
            ]} />
          </div>
        </div>
        <div className="panel" style={{ borderColor: 'color-mix(in srgb, var(--err) 40%, transparent)' }}>
          <div className="panel-hd" style={{ borderBottomColor: 'color-mix(in srgb, var(--err) 30%, transparent)' }}>
            <span className="panel-title tone-err">Danger zone</span>
          </div>
          <div className="panel-bd" style={{ padding: 'var(--pad)' }}>
            <ConfirmInput expect="ingest-gw" actionLabel="Delete"
                          onConfirm={() => push('err', 'Service ingest-gw deleted')} />
          </div>
        </div>
      </section>

      {/* panel states */}
      <section className="sheet-section" data-screen-label="Sheet: Panel states">
        <h2 className="sheet-h">Panel states</h2>
        <Tabs boxed value={stateTab} onChange={setStateTab} items={[
          { id: 'skeleton', label: 'Skeleton' },
          { id: 'loading', label: 'Loading' },
          { id: 'empty', label: 'Empty' },
          { id: 'err', label: 'Error' },
        ]} />
        <div className="panel" style={{ minHeight: 130 }}>
          <div className="panel-bd" style={{ display: 'flex', flexDirection: 'column', justifyContent: 'center' }}>
            {stateTab === 'skeleton' ? <PanelState kind="skeleton" /> : null}
            {stateTab === 'loading' ? <PanelState kind="loading" title="Fetching pods…" sub="kubectl get pods -n prod" /> : null}
            {stateTab === 'empty' ? <PanelState kind="empty" title="No incidents" sub="Nothing matching your filters in the last 7d."
                action={<Btn variant="ghost" onClick={() => push('info', 'Filters cleared')}>Clear filters</Btn>} /> : null}
            {stateTab === 'err' ? <PanelState kind="err" title="Failed to load events" sub="503 from events-api · retried 3×"
                action={<Btn onClick={() => push('ok', 'Reconnected to events-api')}>Retry</Btn>} /> : null}
          </div>
        </div>
      </section>

      {/* split pane */}
      <section className="sheet-section" data-screen-label="Sheet: Split pane">
        <h2 className="sheet-h">Split pane (drag the gutter)</h2>
        <div className="panel" style={{ height: 180 }}>
          <SplitPane initial={42}
            left={
              <div style={{ padding: '3px 0' }}>
                <Tree defaultOpen={['cfg']} nodes={[
                  { id: 'cfg', label: 'config/', children: [
                    { id: 'cfg-d', label: 'deploy.yaml', tone: 'accent' },
                    { id: 'cfg-f', label: 'flags.toml' },
                    { id: 'cfg-r', label: 'limits.toml' },
                  ]},
                ]} />
              </div>
            }
            right={
              <div style={{ padding: 'var(--pad)' }}>
                <DefList rows={[
                  { k: 'file', v: 'deploy.yaml' },
                  { k: 'size', v: '2.1 KiB' },
                  { k: 'changed', v: '4m ago', tone: 'warn' },
                  { k: 'by', v: 'mira@' },
                ]} />
              </div>
            }
          />
        </div>
      </section>

      {/* diff viewer */}
      <section className="sheet-section" data-screen-label="Sheet: Diff viewer">
        <h2 className="sheet-h">Diff viewer</h2>
        <div className="panel">
          <div className="panel-bd">
            <DiffViewer file="deploy/ingest-gw/deploy.yaml" lines={DIFF_SAMPLE} />
          </div>
        </div>
      </section>

    </React.Fragment>
  );
}

window.SheetExtras = SheetExtras;
