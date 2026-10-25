import Link from '@docusaurus/Link';
import Layout from '@theme/Layout';

export default function Home() {
  return (
    <Layout title="Nomyr documentation">
      <main className="container margin-vert--lg">
        <h1>Nomyr documentation</h1>
        <p>Documentation for the Phase 0 foundation scaffold and synthetic local demo.</p>
        <p><Link className="button button--primary" to="/docs/">Read the documentation</Link></p>
      </main>
    </Layout>
  );
}
