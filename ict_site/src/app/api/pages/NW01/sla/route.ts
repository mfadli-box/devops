import { BACKEND_URL } from "@/lib/backend";
import { streamToResponse, getProxyHeaders, handleGlobalError } from "@/lib/apiproxy";

export async function GET(request: Request) {
  try {
    const url = new URL(request.url);
    const Plimit = (url.searchParams.get("limit") || "").trim();
    const Poffset = (url.searchParams.get("offset") || "").trim();
    const query = (Plimit ? `?limit=${encodeURIComponent(Plimit)}` : "?limit=10") +
        (Poffset ? `&offset=${encodeURIComponent(Poffset)}` : "&offset=0");

    const response = await fetch(`${BACKEND_URL}/rest/pages/NW01/sla${query}`, {
      method: "GET",
      headers: getProxyHeaders(request),
    });
    return streamToResponse(response);
  } catch (error) {
    return handleGlobalError(error);
  }
}
