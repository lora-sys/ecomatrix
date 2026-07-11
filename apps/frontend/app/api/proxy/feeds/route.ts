import { NextResponse } from "next/server";
export const dynamic = "force-dynamic";
export async function GET(req: Request) {
  const u = new URL(req.url);
  const limit = u.searchParams.get("limit") ?? "50";
  const backend = `${process.env.NEXT_PUBLIC_BACKEND_URL}/v1/feeds?limit=${limit}`;
  const r = await fetch(backend, { cache: "no-store" });
  return new NextResponse(await r.text(), {
    status: r.status,
    headers: { "content-type": "application/json" },
  });
}
