import React from "react";
import {CloseButton} from "@/components/Buttons";

export default function ModalLayout({children, onClose, title, defaultLayout = true}: { children: React.ReactNode, onClose: () => void, title: string, defaultLayout?: boolean }) {
    if (!defaultLayout) return (<div onClick={onClose} className="fixed inset-0 z-50">{children}</div>);
    return (<div onClick={onClose} className="fixed inset-0 flex justify-center items-center z-50 bg-black/50">
        <div onClick={(e) => e.stopPropagation()} className="p-6 bg-white custom-shadow-m border min-w-90">
            <div className="flex flex-col items-start">
                <div className="flex justify-between mb-8 w-full">
                    <span className="uppercase font-wide font-bold font-8xl pr-6">{title}</span>
                    <CloseButton onClick={onClose} />
                </div>
                {children}
            </div>
        </div>
    </div>);
}
