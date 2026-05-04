import React from "react";
import Link from "next/link";

export default function Section({children, title, href}: {children: React.ReactNode, title: string, href: string}) {
    return (<section className="flex flex-col gap-2 m-4 sm:m-6">
        <Link className="uppercase font-wide text-xl font-bold hover:text-black-hover" href={href}>{title + " >"}</Link>
        {children}
    </section>);
}
