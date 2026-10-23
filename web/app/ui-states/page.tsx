import {DataState, dataStateNames, NothingEvaluable, type DataStateName} from "../../components/data-states/data-state";

const examples: Record<DataStateName, {title: string; message: string}> = {
  loading: {title: "Loading example", message: "A request is in progress."},
  empty: {title: "No findings", message: "The example completed and found no matching findings."},
  partial: {title: "Incomplete result", message: "Some sources did not return records; this result is incomplete."},
  "permission-denied": {title: "Permission denied", message: "The viewer is not permitted to view the requested data."},
  stale: {title: "Possibly out of date", message: "The displayed information may not reflect the current source state."},
  "source-unavailable": {title: "Source unavailable", message: "A contributing source could not be reached."},
  error: {title: "Request failed", message: "The example request did not complete successfully."},
};

export default function UIStatesPage() {
  return <main>
    <header className="topbar"><a href="/">Nomyr</a><a href="/identities/demo-service-account/">Static identity route example</a></header>
    <div className="eyebrow">Component state harness · synthetic examples</div>
    <h1>Reusable data states</h1>
    <p className="lede">Seven reusable data states, plus a distinct nothing-evaluable presentation. These examples contain no provider data or operational controls.</p>
    <div className="state-grid" aria-label="Seven reusable data state examples">
      {dataStateNames.map((state) => <DataState key={state} state={state} title={examples[state].title} live={state === "loading"}>{examples[state].message}</DataState>)}
      <NothingEvaluable title="No evaluation was possible">Insufficient evidence or inputs were available to evaluate this example. This is distinct from an evaluated result with no findings.</NothingEvaluable>
    </div>
    <p className="watermark">Demo examples only</p>
  </main>;
}
