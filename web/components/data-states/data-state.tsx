import {useId, type ReactNode} from "react";

export const dataStateNames = ["loading", "empty", "partial", "permission-denied", "stale", "source-unavailable", "error"] as const;
export type DataStateName = typeof dataStateNames[number];

const labels: Record<DataStateName, string> = {
  loading: "Loading", empty: "Empty", partial: "Partial", "permission-denied": "Permission denied",
  stale: "Stale", "source-unavailable": "Source unavailable", error: "Error",
};

export function DataState({state, title, children, live = false}: {
  state: DataStateName;
  title: string;
  children: ReactNode;
  live?: boolean;
}) {
  const id = useId();
  const role = state === "loading" ? "status" : state === "error" ? "alert" : undefined;
  return <section className="state-card" data-state={state} aria-labelledby={`${id}-title`} role={role} aria-live={live ? "polite" : undefined}>
    <span className="state-label">{labels[state]}</span>
    <h2 id={`${id}-title`}>{title}</h2>
    <div>{children}</div>
  </section>;
}

export function NothingEvaluable({title, children}: {title: string; children: ReactNode}) {
  const id = useId();
  return <section className="state-card nothing-evaluable" data-state="nothing-evaluable" aria-labelledby={`${id}-title`}>
    <span className="state-label">Nothing evaluable</span>
    <h2 id={`${id}-title`}>{title}</h2>
    <div>{children}</div>
  </section>;
}
