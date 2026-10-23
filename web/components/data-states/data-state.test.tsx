import {cleanup, render, screen} from "@testing-library/react";
import axe from "axe-core";
import {afterEach, describe, expect, it} from "vitest";

import {DataState, dataStateNames, NothingEvaluable} from "./data-state";

afterEach(cleanup);

describe("DataState", () => {
  it("renders all seven separate states with meaningful labels", () => {
    const {container} = render(<>
      {dataStateNames.map((state) => <DataState key={state} state={state} title={`${state} example`}>Example detail</DataState>)}
    </>);
    expect(dataStateNames).toEqual(["loading", "empty", "partial", "permission-denied", "stale", "source-unavailable", "error"]);
    expect(container.querySelectorAll("[data-state]")).toHaveLength(7);
    expect(screen.getByRole("alert").getAttribute("data-state")).toBe("error");
    expect(screen.getByRole("status").getAttribute("data-state")).toBe("loading");
    expect(screen.getByText("Permission denied")).toBeTruthy();
    expect(screen.getByText("Source unavailable")).toBeTruthy();
    expect(screen.getByText("Partial")).toBeTruthy();
  });

  it("renders nothing evaluable separately from an empty result", () => {
    const {container} = render(<><DataState state="empty" title="No findings">Evaluated with none found.</DataState><NothingEvaluable title="Could not evaluate">Inputs were insufficient.</NothingEvaluable></>);
    expect(container.querySelector('[data-state="empty"]')).toBeTruthy();
    expect(container.querySelector('[data-state="nothing-evaluable"]')).toBeTruthy();
    expect(screen.getByText("Nothing evaluable")).toBeTruthy();
  });

  it("has no axe violations in the seven-state example", async () => {
    const {container} = render(<main><h1>State examples</h1>{dataStateNames.map((state) => <DataState key={state} state={state} title={`${state} example`}>Example detail</DataState>)}<NothingEvaluable title="No evaluation">Evidence is insufficient.</NothingEvaluable></main>);
    const results = await axe.run(container, {rules: {"color-contrast": {enabled: false}}});
    expect(results.violations).toEqual([]);
  });
});
