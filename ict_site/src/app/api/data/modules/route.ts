import { BACKEND_URL } from "@/lib/backend";
import { streamToResponse, getProxyHeaders, handleGlobalError } from "@/lib/apiproxy";

export async function GET(request: Request) {
  try {
    const url = new URL(request.url);
    const companyId = (url.searchParams.get("company_id") || "").trim();
    const query = companyId ? `?company_id=${encodeURIComponent(companyId)}` : "";

    const response = await fetch(`${BACKEND_URL}/rest/module${query}`, {
      headers: getProxyHeaders(request),
    });
    return streamToResponse(response);
  } catch (error) {
    return handleGlobalError(error);
  }
}
