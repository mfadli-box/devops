import { BACKEND_URL } from "@/lib/backend";
import { streamToResponse, getProxyHeaders, handleGlobalError } from "@/lib/apiproxy";

export async function GET(request: Request) {
  try {
    const url = new URL(request.url);

    const startDate = (url.searchParams.get("start_date") || "").trim();
    const endDate = (url.searchParams.get("end_date") || "").trim();
    const page = (url.searchParams.get("page") || "").trim();
    const limit = (url.searchParams.get("limit") || "").trim();

    const query = `?start_date=${encodeURIComponent(startDate)}` +
                  `&end_date=${encodeURIComponent(endDate)}` +
                  `&page=${page ? encodeURIComponent(page) : "1"}` +
                  `&limit=${limit ? encodeURIComponent(limit) : "10"}`;

    const response = await fetch(`${BACKEND_URL}/rest/pages/NW02/sla${query}`, {
      method: "GET",
      headers: getProxyHeaders(request),
    });
    return streamToResponse(response);
  } catch (error) {
    return handleGlobalError(error);
  }
}
