import { BACKEND_URL } from "@/lib/backend";
import { streamToResponse, getProxyHeaders, handleGlobalError } from "@/lib/apiproxy";

export async function GET(request: Request) {
  try {
    const url = new URL(request.url);

    const monitorId = (url.searchParams.get("monitor_id") || "").trim();
    const domain = (url.searchParams.get("domain") || "").trim();
    const page = (url.searchParams.get("page") || "").trim();
    const limit = (url.searchParams.get("limit") || "").trim();

    const query = `?monitor_id=${encodeURIComponent(monitorId)}` +
                  `&domain=${encodeURIComponent(domain)}` +
                  `&page=${page ? encodeURIComponent(page) : "1"}` +
                  `&limit=${limit ? encodeURIComponent(limit) : "10"}`;

    const response = await fetch(`${BACKEND_URL}/rest/pages/NW02/logs${query}`, {
      method: "GET",
      headers: getProxyHeaders(request),
    });
    return streamToResponse(response);
  } catch (error) {
    return handleGlobalError(error);
  }
}
