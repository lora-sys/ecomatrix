import { NextResponse } from "next/server";
export const dynamic = "force-dynamic";
export async function GET() {
  const r = await fetch(`${process.env.NEXT_PUBLIC_BACKEND_URL}/v1/metrics`, { cache: "no-store" });
  return new NextResponse(await r.text(), {
    status: r.status,
    headers: { "content-type": "application/json" },
  });
}
