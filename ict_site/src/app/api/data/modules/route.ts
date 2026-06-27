import { NextResponse } from "next/server";
import { BACKEND_URL } from "@/lib/backend";

export async function GET(request: Request) {
  try {
    const authorization = request.headers.get("authorization") || "";
    const url = new URL(request.url);
    const companyId = (url.searchParams.get("company_id") || "").trim();
    const query = companyId ? `?company_id=${encodeURIComponent(companyId)}` : "";

    const response = await fetch(`${BACKEND_URL}/rest/module${query}`, {
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
