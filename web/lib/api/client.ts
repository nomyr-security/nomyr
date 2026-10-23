import type { paths } from "./openapi";

type Methods<Path extends keyof paths> = Extract<keyof paths[Path], string>;
type ResponseFor<Path extends keyof paths, Method extends Methods<Path>> =
  paths[Path][Method] extends { responses: infer Responses }
    ? Responses[keyof Responses] extends { content: infer Content }
      ? Content[keyof Content]
      : never
    : never;

export async function request<Path extends keyof paths, Method extends Methods<Path>>(
  path: Path,
  method: Method,
  init: RequestInit = {},
): Promise<ResponseFor<Path, Method>> {
  const response = await fetch(`/api/v1${path}`, { ...init, method });
  if (!response.ok) {
    throw new Error(`Nomyr API request failed with HTTP ${response.status}`);
  }
  return (await response.json()) as ResponseFor<Path, Method>;
}
