const states = ["Loading", "Data", "Empty", "No access", "Error", "Partial", "Unknown"];

export default function HomePage() {
  return (
    <main>
      <div className="eyebrow">Non-human identity security</div>
      <h1><img className="brand-lockup" src="/brand/nomyr-lockup-on-dark.svg" alt="Nomyr" width="341" height="96" /></h1>
      <p className="lede">From discovery and governance to identity lifecycle and AI agent access.</p>
      <section aria-labelledby="status-heading">
        <h2 id="status-heading">Local development demo</h2>
        <div className="status-grid">
          <article><strong>Explore</strong><span>A local interface with synthetic identity data</span></article>
          <article><strong>Demo scope</strong><span>No live provider connections or security actions</span></article>
        </div>
      </section>
      <section aria-labelledby="states-heading">
        <h2 id="states-heading">Required data states</h2>
        <ul>{states.map((state) => <li key={state}>{state}</li>)}</ul>
      </section>
      <p className="watermark">Demo data only</p>
    </main>
  );
}
