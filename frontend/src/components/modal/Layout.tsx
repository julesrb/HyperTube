import React from "react";
import {CloseButton} from "@/components/Buttons";

export default function ModalLayout({children, onClose, title}: {children: React.ReactNode, onClose: () => void, title: string}) {
    return (<div onClick={onClose} className="fixed inset-0 flex justify-center items-center z-50 bg-black/50">
        <div onClick={(e) => e.stopPropagation()} className="p-6 bg-white custom-shadow-m border min-w-9/10 sm:min-w-90 max-w-9/10 sm:max-w-none">
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
