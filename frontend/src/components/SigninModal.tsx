"use client";

import { useModal } from "@/context/ModalContext";
import ModalLayout from "@/components/LayoutModal";
import React from "react";
import Input from "@/components/Input";
import Button from "@/components/Button";
import CrossIcon from "@/components/icon/CrossIcon";

export default function SigninModal() {
    const {activeModal, closeModal,} = useModal();

    if (activeModal !== "signin")
        return null;

    return (
        <ModalLayout onClose={closeModal} defaultLayout={false}>
            <div className="absolute top-0 right-1/10 flex gap-2 p-2 bg-white custom-shadow-s custom-border" onClick={(e) => e.stopPropagation()}>
                <button onClick={closeModal}>
                    <CrossIcon />
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