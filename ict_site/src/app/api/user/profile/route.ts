import { BACKEND_URL } from "@/lib/backend";
import { streamToResponse, getProxyHeaders, handleGlobalError } from "@/lib/apiproxy";

export async function GET(request: Request) {
  try {
    const response = await fetch(`${BACKEND_URL}/rest/user/profile`, {
      method: "GET",
      headers: getProxyHeaders(request),
    });
    return streamToResponse(response);
  } catch (error) {
    return handleGlobalError(error);
  }
}

export async function PUT(request: Request) {
  try {
    const body = await request.json();
    const response = await fetch(`${BACKEND_URL}/rest/user/profile`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
        ...getProxyHeaders(request),
      },
      body: JSON.stringify(body),
    });
    return streamToResponse(response);
  } catch (error) {
    return handleGlobalError(error);
  }
}
