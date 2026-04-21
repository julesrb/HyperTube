"use client";

import { useModal } from "@/context/ModalContext";
import ModalLayout from "@/components/modal/Layout";
import React from "react";
import Input from "@/components/Input";
import {Button} from "@/components/Button";

export default function Register() {
    const { activeModal, closeModal, } = useModal();

    if (activeModal.type !== "register")
        return null;

    return (
        <ModalLayout onClose={closeModal} title="Join Hypertube">
            <Input type="email" placeholder="Email"></Input>

            <div className="flex gap-2">
                <Input type="firstname" placeholder="Firstname"></Input>
                <Input type="lastname" placeholder="Lastname"></Input>
            </div>

            <Input type="username" placeholder="Username" className={"max-w-2/3"}></Input>
            <Input type="password" placeholder="Password" className={"max-w-2/3"}></Input>

            <Button className="h-8 mt-2">Sign Up</Button>
        </ModalLayout>
    );
}
