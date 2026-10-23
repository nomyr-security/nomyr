import {cleanup, fireEvent, render, screen} from "@testing-library/react";
import {afterEach, describe, expect, it} from "vitest";

import {Tabs} from "./tabs";

afterEach(cleanup);

describe("Tabs", () => {
  it("supports selection and arrow-key navigation", () => {
    render(<Tabs label="Example sections" items={[
      {id: "one", label: "One", content: <p>First panel</p>},
      {id: "two", label: "Two", content: <p>Second panel</p>},
    ]} />);
    const first = screen.getByRole("tab", {name: "One"});
    const second = screen.getByRole("tab", {name: "Two"});
    fireEvent.keyDown(first, {key: "ArrowRight"});
    expect(document.activeElement).toBe(second);
    expect(second.getAttribute("aria-selected")).toBe("true");
    expect(screen.getByRole("tabpanel", {name: "Two"}).hasAttribute("hidden")).toBe(false);
    fireEvent.click(first);
    expect(first.getAttribute("aria-selected")).toBe("true");
  });
});
