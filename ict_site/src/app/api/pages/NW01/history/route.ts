import { BACKEND_URL } from "@/lib/backend";
import { streamToResponse, getProxyHeaders, handleGlobalError } from "@/lib/apiproxy";

export async function GET(request: Request) {
  try {
    const url = new URL(request.url);
    const Pip = (url.searchParams.get("ip") || "").trim();
    const Pdate = (url.searchParams.get("date") || "").trim();
    const query = (Pip ? `?ip=${encodeURIComponent(Pip)}` : "?ip=") +
        (Pdate ? `&date=${encodeURIComponent(Pdate)}` : "&date="+ new Date().toISOString().split("T")[0]);

    const response = await fetch(`${BACKEND_URL}/rest/pages/NW01/history${query}`, {
      method: "GET",
      headers: getProxyHeaders(request),
    });
    return streamToResponse(response);
  } catch (error) {
    return handleGlobalError(error);
  }
}
