import { BACKEND_URL } from "@/lib/backend";
import { streamToResponse, getProxyHeaders, handleGlobalError } from "@/lib/apiproxy";

export async function GET(request: Request) {
  try {
    const authorization = request.headers.get("authorization") || "";
    const endpoint = authorization ? `${BACKEND_URL}/rest/user/company` : `${BACKEND_URL}/rest/company`;
    const response = await fetch(endpoint, {
      headers: getProxyHeaders(request),
    });
    return streamToResponse(response);
  } catch (error) {
    return handleGlobalError(error);
  }
}
