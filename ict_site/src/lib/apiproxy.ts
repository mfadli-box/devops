import { NextResponse } from "next/server";

export const streamToResponse = (response: Response) => {
  if (!response.ok) {
    return handleBackendError(response.status);
  }

  if (!response.body) {
    return NextResponse.json({}, { status: response.status });
  }

  return new NextResponse(response.body, {
    status: response.status,
    headers: {
      "Content-Type": "application/json",
      "X-Accel-Buffering": "no",
    },
  });
};

export const getProxyHeaders = (request: Request) => {
  return {
    Authorization: request.headers.get("authorization") || "",
    Cookie: request.headers.get("cookie") || "",
  };
};

export const handleGlobalError = (error: unknown) => {
  console.error("[API Proxy Error]:", error);
  
  return NextResponse.json(
    { error: "Gagal menghubungi sistem layanan internal." }, 
    { status: 504 }
  );
};

const handleBackendError = (status: number) => {
  let message = "Terjadi kesalahan pada sistem layanan.";
  
  switch (status) {
    case 400:
      message = "Format permintaan data tidak valid.";
      break;
    case 401:
      message = "Sesi Anda telah berakhir. Silakan login kembali.";
      break;
    case 403:
      message = "Anda tidak memiliki hak akses untuk melakukan aksi ini.";
      break;
    case 404:
      message = "Data atau layanan tidak ditemukan.";
      break;
    case 500:
      message = "Layanan backend mengalami gangguan internal.";
      break;
  }

  return NextResponse.json({ error: message }, { status });
};
