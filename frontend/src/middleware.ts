import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

const COOKIE_NAME = "gotest_token";

// Gate autentikasi global: seluruh aplikasi dilindungi login. User yang belum
// login di-redirect ke /login (tujuan asal dibawa di query param `redirect`).
// Validasi JWT sesungguhnya tetap dilakukan backend di setiap API call;
// middleware ini hanya gate UX cepat berdasarkan keberadaan cookie session,
// supaya user harus login di "pintu depan" sebelum mengakses halaman apa pun.
export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;
  const token = request.cookies.get(COOKIE_NAME)?.value;

  if (!token) {
    const loginUrl = new URL("/login", request.url);
    if (pathname && pathname !== "/") {
      loginUrl.searchParams.set("redirect", pathname);
    }
    return NextResponse.redirect(loginUrl);
  }

  return NextResponse.next();
}

export const config = {
  // Jalankan untuk semua path kecuali asset Next.js, favicon, dan halaman login.
  matcher: ["/((?!_next/static|_next/image|favicon.ico|login).*)"],
};
