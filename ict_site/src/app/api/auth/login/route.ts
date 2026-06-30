import { NextResponse } from "next/server";
import { BACKEND_URL } from "@/lib/backend";
import { handleGlobalError } from "@/lib/apiproxy";

export async function POST(request: Request) {
  try {
    const body = await request.json();

    const response = await fetch(`${BACKEND_URL}/rest/user/login`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(body),
    });

    const responseBody = await response.json();
    if (!response.ok) {
      return NextResponse.json(responseBody, { status: response.status });
    }

    const setCookie = response.headers.get("set-cookie");
    const nextResponse = NextResponse.json(responseBody.data ?? responseBody, { status: response.status });
    if (setCookie) {
      nextResponse.headers.set("set-cookie", setCookie);
    }
    return nextResponse;
  } catch (error) {
    return handleGlobalError(error);
  }
}
