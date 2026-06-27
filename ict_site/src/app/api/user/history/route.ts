import { NextResponse } from "next/server";
import { BACKEND_URL } from "@/lib/backend";

export async function GET(request: Request) {
  try {
    const authorization = request.headers.get("authorization") || "";
    const cookie = request.headers.get("cookie") || "";

    const response = await fetch(`${BACKEND_URL}/rest/user/history`, {
      method: "GET",
      headers: {
        Authorization: authorization,
        Cookie: cookie,
      },
    });

    const data = await response.json();
    return NextResponse.json(data, { status: response.status });
  } catch {
    return NextResponse.json({ error: "Gagal menghubungi sistem layanan." }, { status: 500 });
  }
}
