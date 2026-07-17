import useModal from "@/contexts/ModalContext";
import useAuth from "@/contexts/AuthContext";
import Link from "next/link";
import React from "react";

export default function LinkLoginRequired({children, href, className}: {children?: React.ReactNode, href: string, className?: string}) {
    const {openModal} = useModal();
    const {user, setCallbackUrl} = useAuth();

    return (<Link href={href} onClick={e => {
        if (!user) {
            e.preventDefault();
            setCallbackUrl(href.toString());
            openModal({type: "signin"});
            return;
        }
    }} className={className} >
        {children}
    </Link>);
}
