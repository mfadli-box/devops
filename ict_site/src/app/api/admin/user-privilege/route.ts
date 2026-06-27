import { NextResponse } from "next/server";
import { BACKEND_URL } from "@/lib/backend";

const toJsonResponse = async (response: Response) => {
  const raw = await response.text();
  const data = raw ? (() => {
    try {
      return JSON.parse(raw);
    } catch {
      return { error: raw };
    }
  })() : {};

  return NextResponse.json(data, { status: response.status });
};

export async function GET(request: Request) {
  try {
    const authorization = request.headers.get("authorization") || "";
    const cookie = request.headers.get("cookie") || "";
    const url = new URL(request.url);
    const userCompanyId = (url.searchParams.get("user_company_id") || "").trim();
    const query = userCompanyId ? `?user_company_id=${encodeURIComponent(userCompanyId)}` : "";

    const response = await fetch(`${BACKEND_URL}/rest/admin/user-privilege${query}`, {
      method: "GET",
      headers: { Authorization: authorization, Cookie: cookie },
    });
    return toJsonResponse(response);
  } catch {
    return NextResponse.json({ error: "Gagal menghubungi sistem layanan." }, { status: 500 });
  }
}

export async function POST(request: Request) {
  try {
    const authorization = request.headers.get("authorization") || "";
    const cookie = request.headers.get("cookie") || "";
    const body = await request.json();

    const response = await fetch(`${BACKEND_URL}/rest/admin/user-privilege`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: authorization,
        Cookie: cookie,
      },
      body: JSON.stringify(body),
    });

    return toJsonResponse(response);
  } catch {
    return NextResponse.json({ error: "Gagal menghubungi sistem layanan." }, { status: 500 });
  }
}

export async function PUT(request: Request) {
  try {
    const authorization = request.headers.get("authorization") || "";
    const cookie = request.headers.get("cookie") || "";
    const body = await request.json();

    if (!body?.id) {
      return NextResponse.json({ error: "ID modul pengguna wajib diisi." }, { status: 400 });
    }

    const response = await fetch(`${BACKEND_URL}/rest/admin/user-privilege/${body.id}`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
        Authorization: authorization,
        Cookie: cookie,
      },
      body: JSON.stringify(body),
    });

    return toJsonResponse(response);
  } catch {
    return NextResponse.json({ error: "Gagal menghubungi sistem layanan." }, { status: 500 });
  }
}
