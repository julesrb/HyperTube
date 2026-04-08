"use client";

import { useModal } from "@/context/ModalContext";
import ModalLayout from "@/components/LayoutModal";
import Image from "next/image";
import React from "react";
import Input from "@/components/Input";
import Button from "@/components/Button";

export default function SigninModal() {
    const {activeModal, closeModal,} = useModal();

    if (activeModal !== "signin")
        return null;

    return (
        <ModalLayout onClose={closeModal} defaultLayout={false}>
            <div className="absolute top-0 right-1/10 flex gap-2 p-2 bg-white custom-shadow-s custom-border" onClick={(e) => e.stopPropagation()}>
                <button onClick={closeModal}>
                    <Image className="custom-icon-black" src="/icons/cross.svg" alt="cross" width={30} height={30}/>
                </button>
                <Input type="email" placeholder="Email"></Input>
                <Input type="password" placeholder="Password"></Input>

                <div className="flex" style={{alignItems: "flex-end", marginRight: "8px"}}>
                    <Button className="h-8">Sign In</Button>
                </div>
            </div>
        </ModalLayout>
    );
}