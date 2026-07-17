import {Link} from "@/i18n/navigation";
import React from "react";

export default function Section({children, title, href}: {children: React.ReactNode, title: string, href: string}) {
    return (<section className="flex flex-col gap-2">
        <Link className="uppercase font-wide text-xl font-bold hover:text-black-hover" href={href}>{title + " >"}</Link>
        {children}
    </section>);
}
