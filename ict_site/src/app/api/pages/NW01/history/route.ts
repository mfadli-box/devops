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
    const Pip = (url.searchParams.get("ip") || "").trim();
    const Pdate = (url.searchParams.get("date") || "").trim();
    const query = (Pip ? `?ip=${encodeURIComponent(Pip)}` : "?ip=") +
        (Pdate ? `&date=${encodeURIComponent(Pdate)}` : "&date="+ new Date().toISOString().split("T")[0]);

    const response = await fetch(`${BACKEND_URL}/rest/pages/NW01/history${query}`, {
      method: "GET",
      headers: { Authorization: authorization, Cookie: cookie },
    });
    return toJsonResponse(response);
  } catch {
    return NextResponse.json({ error: "Gagal menghubungi sistem layanan." }, { status: 500 });
  }
}
