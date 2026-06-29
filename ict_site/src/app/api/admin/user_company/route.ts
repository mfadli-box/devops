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
    const userId = (url.searchParams.get("user_id") || "").trim();
    const query = userId ? `?user_id=${encodeURIComponent(userId)}` : "";

    const response = await fetch(`${BACKEND_URL}/rest/admin/user-company${query}`, {
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

    const response = await fetch(`${BACKEND_URL}/rest/admin/user-company`, {
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
      return NextResponse.json({ error: "ID perusahaan pengguna wajib diisi." }, { status: 400 });
    }

    const response = await fetch(`${BACKEND_URL}/rest/admin/user-company/${body.id}`, {
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
