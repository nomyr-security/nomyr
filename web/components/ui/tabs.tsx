"use client";

import {useId, useState, type KeyboardEvent, type ReactNode} from "react";

export type TabItem = {id: string; label: string; content: ReactNode};

export function Tabs({items, label}: {items: TabItem[]; label: string}) {
  const id = useId();
  const [selected, setSelected] = useState(0);
  function onKeyDown(event: KeyboardEvent<HTMLButtonElement>) {
    const direction = event.key === "ArrowRight" ? 1 : event.key === "ArrowLeft" ? -1 : 0;
    if (!direction) return;
    event.preventDefault();
    const next = (selected + direction + items.length) % items.length;
    setSelected(next);
    document.getElementById(`${id}-tab-${next}`)?.focus();
  }

  return <>
    <div className="tablist" role="tablist" aria-label={label}>
      {items.map((item, index) => <button
        key={item.id}
        id={`${id}-tab-${index}`}
        className="tab"
        type="button"
        role="tab"
        aria-selected={selected === index}
        aria-controls={`${id}-panel-${index}`}
        tabIndex={selected === index ? 0 : -1}
        onClick={() => setSelected(index)}
        onKeyDown={onKeyDown}
      >{item.label}</button>)}
    </div>
    {items.map((item, index) => <section
      key={item.id}
      id={`${id}-panel-${index}`}
      className="tab-panel"
      role="tabpanel"
      aria-labelledby={`${id}-tab-${index}`}
      hidden={selected !== index}
      tabIndex={0}
    >{item.content}</section>)}
  </>;
}
