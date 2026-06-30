import { NextResponse } from "next/server";
import { BACKEND_URL } from "@/lib/backend";
import { streamToResponse, getProxyHeaders, handleGlobalError } from "@/lib/apiproxy";

export async function GET(request: Request) {
  try {
    const url = new URL(request.url);
    const userCompanyId = (url.searchParams.get("user_company_id") || "").trim();
    const query = userCompanyId ? `?user_company_id=${encodeURIComponent(userCompanyId)}` : "";
    const response = await fetch(`${BACKEND_URL}/rest/admin/user-privilege${query}`, {
      method: "GET",
      headers: getProxyHeaders(request),
    });
    return streamToResponse(response);
  } catch (error) {
    return handleGlobalError(error);
  }
}

export async function POST(request: Request) {
  try {
    const body = await request.json();
    const response = await fetch(`${BACKEND_URL}/rest/admin/user-privilege`, {
      method: "POST",
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

export async function PUT(request: Request) {
  try {
    const body = await request.json();
    if (!body?.id) {
      return NextResponse.json({ error: "ID modul pengguna wajib diisi." }, { status: 400 });
    }

    const response = await fetch(`${BACKEND_URL}/rest/admin/user-privilege/${body.id}`, {
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
