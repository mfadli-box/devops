import { NextResponse } from "next/server";
import { BACKEND_URL } from "@/lib/backend";

export async function PUT(request: Request) {
  try {
    const body = await request.json();
    const authorization = request.headers.get("authorization") || "";
    const cookie = request.headers.get("cookie") || "";

    const response = await fetch(`${BACKEND_URL}/rest/user/password`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
        Authorization: authorization,
        Cookie: cookie,
      },
      body: JSON.stringify(body),
    });

    const data = await response.json();
    return NextResponse.json(data, { status: response.status });
  } catch {
    return NextResponse.json({ error: "Gagal menghubungi sistem layanan." }, { status: 500 });
  }
}
