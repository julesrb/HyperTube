import createMiddleware from "next-intl/middleware";
import {routing} from "@/i18n/request";

export default createMiddleware(routing);

// noinspection JSUnusedGlobalSymbols
export const config = {matcher: "/((?!api|trpc|_next|_vercel|.*\\..*).*)"};
