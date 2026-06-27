import { NextResponse } from "next/server";
import { BACKEND_URL } from "@/lib/backend";

export async function GET(request: Request) {
  try {
    const authorization = request.headers.get("authorization") || "";
    const endpoint = authorization ? `${BACKEND_URL}/rest/user/company` : `${BACKEND_URL}/rest/company`;
    const response = await fetch(endpoint, {
      headers: {
        Authorization: authorization,
        Cookie: request.headers.get("cookie") || "",
      },
    });
    const data = await response.json();
    return NextResponse.json(data, { status: response.status });
  } catch (err) {
    return NextResponse.json({ error: "Gagal menghubungi sistem layanan." }, { status: 500 });
  }
}
