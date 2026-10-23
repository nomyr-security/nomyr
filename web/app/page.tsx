import {getMessages} from "../lib/i18n/messages";

export default function HomePage() {
  const copy = getMessages();
  return (
    <main>
      <div className="eyebrow">{copy.demoLabel}</div>
      <h1><img className="brand-lockup" src="/brand/nomyr-lockup-on-dark.svg" alt="Nomyr" width="341" height="96" /></h1>
      <p className="lede">{copy.syntheticOnly}</p>
      <section aria-labelledby="status-heading">
        <h2 id="status-heading">Local development demo</h2>
        <div className="status-grid">
          <article><strong>Explore</strong><span>A local interface with synthetic identity data</span></article>
          <article><strong>Demo scope</strong><span>No live provider connections or security actions</span></article>
        </div>
      </section>
      <p><a className="button" href="/ui-states/">Explore component states</a></p>
      <p className="watermark">Demo data only</p>
    </main>
  );
}
