"use client";

import {useSearchParams} from "next/navigation";

export function AsOf() {
  const searchParams = useSearchParams();
  return <p>As of: {searchParams.get("as_of") ?? "latest synthetic snapshot"}</p>;
}
