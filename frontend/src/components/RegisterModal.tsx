"use client";

import { useModal } from "@/context/ModalContext";
import ModalLayout from "@/components/LayoutModal";
import React from "react";
import Image from "next/image";
import Input from "@/components/Input";
import Button from "@/components/Button";

export default function RegisterModal() {
    const { activeModal, closeModal, } = useModal();

    if (activeModal !== "register")
        return null;

    return (
        <ModalLayout onClose={closeModal} defaultLayout={true}>
            <div className="flex justify-between mb-2">
                <h4 className="uppercase">Join Hypertube</h4>
                <button onClick={closeModal}>
                    <Image className="custom-icon-black" src="/icons/cross.svg" alt="cross" width={30} height={30}/>
                </button>
            </div>
            <Input type="email" placeholder="Email"></Input>

            <div className="flex gap-2">
                <Input type="firstname" placeholder="Firstname"></Input>
                <Input type="lastname" placeholder="Lastname"></Input>
            </div>

            <Input type="username" placeholder="Username" className={"max-w-2/3"}></Input>
            <Input type="password" placeholder="Password" className={"max-w-2/3"}></Input>

            <Button className="h-8 max-w-1/4">Sign Up</Button>
        </ModalLayout>
    );
}
