import {Suspense} from "react";

import {AsOf} from "./as-of";

type IdentityPageProps = {params: Promise<{id: string}>};

export function generateStaticParams() {
  return [{id: "demo-service-account"}];
}

export default async function IdentityPage({params}: IdentityPageProps) {
  const {id} = await params;
  return (
    <main>
      <div className="eyebrow">Static-export route spike</div>
      <h1>{id}</h1>
      <Suspense fallback={<p>As of: loading</p>}><AsOf /></Suspense>
      <nav aria-label="Identity detail tabs"><a href="#overview">Overview</a><a href="#evidence">Evidence</a></nav>
      <section id="overview"><h2>Overview</h2><p>This route proves the build-time export shape only. Identity behavior is not implemented.</p></section>
      <section id="evidence"><h2>Evidence</h2><p>No source evidence exists for this synthetic identity.</p></section>
    </main>
  );
}
