import { NextResponse } from "next/server";
import { BACKEND_URL } from "@/lib/backend";
import { getProxyHeaders, handleGlobalError } from "@/lib/apiproxy";

export async function POST(request: Request) {
  try {
    const body = await request.json();
    const response = await fetch(`${BACKEND_URL}/rest/user/logout`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        ...getProxyHeaders(request),
      },
      body: JSON.stringify(body),
    });
    const data = await response.json();
    const nextResponse = NextResponse.json(data, { status: response.status });
    const setCookie = response.headers.get("set-cookie");
    if (setCookie) {
      nextResponse.headers.set("set-cookie", setCookie);
    }
    return nextResponse;
  } catch (error) {
    return handleGlobalError(error);
  }
}
