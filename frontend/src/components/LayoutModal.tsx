"use client";

import React from "react";

export default function ModalLayout({ children, onClose, defaultLayout = true } : { children: React.ReactNode, onClose: () => void, defaultLayout?: boolean }) {
    if (!defaultLayout)
        return (<div onClick={onClose} className="fixed inset-0 z-10" >{children}</div>);
    return (
        <div onClick={onClose} className="fixed bg-black/40 inset-0 flex justify-center items-center z-10">
            <div onClick={(e) => e.stopPropagation()} className="p-6 bg-white custom-shadow-m custom-border w-[380]">
                <div className="flex flex-col gap-2">
                    {children}
                </div>
            </div>
        </div>
    );
}
