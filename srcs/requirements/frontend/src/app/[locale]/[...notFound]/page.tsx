"use client";

import {redirect, usePathname} from "next/navigation";

export default function NotFoundPage() {
    const pathname = usePathname();
    redirect(`/?error=${pathname}`);
}
