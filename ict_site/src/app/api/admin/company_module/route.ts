import { NextResponse } from "next/server";
import { BACKEND_URL } from "@/lib/backend";
import { streamToResponse, getProxyHeaders, handleGlobalError } from "@/lib/apiproxy";

export async function GET(request: Request) {
  try {
    const url = new URL(request.url);
    const companyId = (url.searchParams.get("company_id") || "").trim();
    const query = companyId ? `?company_id=${encodeURIComponent(companyId)}` : "";
    const response = await fetch(`${BACKEND_URL}/rest/admin/company-module${query}`, {
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
    const response = await fetch(`${BACKEND_URL}/rest/admin/company-module`, {
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
      return NextResponse.json({ error: "ID modul perusahaan wajib diisi." }, { status: 400 });
    }

    const response = await fetch(`${BACKEND_URL}/rest/admin/company-module/${body.id}`, {
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
