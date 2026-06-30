import { NextResponse } from "next/server";
import { BACKEND_URL } from "@/lib/backend";
import { streamToResponse, getProxyHeaders, handleGlobalError } from "@/lib/apiproxy";

export async function DELETE(request: Request, { params }: { params: { id: string } }) {
  try {
    const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
    if (!uuidRegex.test(params.id)) {
      return NextResponse.json({ error: "Format ID tidak valid." }, { status: 400 });
    }

    const response = await fetch(`${BACKEND_URL}/rest/pages/NW02/logs/${params.id}`, {
      method: "DELETE",
      headers: getProxyHeaders(request),
    });

    return streamToResponse(response);
  } catch (error) {
    return handleGlobalError(error);
  }
}
