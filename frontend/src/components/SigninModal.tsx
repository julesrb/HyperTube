"use client";

import { useModal } from "@/context/ModalContext";
import ModalLayout from "@/components/LayoutModal";
import React from "react";
import Input from "@/components/Input";
import Button from "@/components/Button";
import CrossIcon from "@/components/icon/CrossIcon";

// export default function SigninModal() {
//     const {activeModal, closeModal,} = useModal();
//
//     if (activeModal !== "signin")
//         return null;
//
//     return (
//         <ModalLayout onClose={closeModal} defaultLayout={false}>
//             <div className="absolute top-0 right-1/10 flex gap-2 p-2 bg-white custom-shadow-s border" onClick={(e) => e.stopPropagation()}>
//                 <button onClick={closeModal}><CrossIcon /></button>
//                 <Input type="email" placeholder="Email"></Input>
//                 <Input type="password" placeholder="Password"></Input>
//
//                 <Button className="h-8">Sign In</Button>
//             </div>
//         </ModalLayout>
//     );
// }

export default function SigninModal() {
    const {activeModal, closeModal,} = useModal();

    if (activeModal !== "signin")
        return null;

    return (
        <ModalLayout onClose={closeModal}>
            <div className="flex justify-between mb-8 w-full">
                <h3 className="uppercase">Sign in</h3>
                <button onClick={closeModal}>
                    <CrossIcon />
                </button>
            </div>
            <Input type="email" placeholder="Email"></Input>
            <Input type="password" placeholder="Password"></Input>
            <div className="relative mb-4">
                <a className="custom-underline text-xs absolute bottom-2">Forgotten?</a>
            </div>
            <Button className="h-8">Sign In</Button>
        </ModalLayout>
    );
}