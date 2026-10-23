import {Suspense} from "react";

import {AsOf} from "./as-of";
import {Tabs} from "../../../components/ui/tabs";

type IdentityPageProps = {params: Promise<{id: string}>};

export function generateStaticParams() {
  return [{id: "demo-service-account"}];
}

export default async function IdentityPage({params}: IdentityPageProps) {
  const {id} = await params;
  return (
    <main>
      <div className="eyebrow">Static route spike · synthetic demo</div>
      <h1>{id}</h1>
      <Suspense fallback={<p role="status">As of: loading query value</p>}><AsOf /></Suspense>
      <p className="notice">Static export route example. Identity retrieval, authorization, and evidence collection are not implemented.</p>
      <Tabs label="Identity detail sections" items={[
        {id: "overview", label: "Overview", content: <><h2>Overview</h2><p>This build-time route demonstrates the exported identity path only. No live identity record is loaded.</p></>},
        {id: "evidence", label: "Evidence", content: <><h2>Evidence</h2><p>No source evidence exists for this synthetic identity.</p></>},
      ]} />
      <p><a href="/ui-states/">View reusable data-state examples</a></p>
    </main>
  );
}
