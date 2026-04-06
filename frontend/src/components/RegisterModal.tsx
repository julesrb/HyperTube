"use client";

import { useModal } from "@/context/ModalContext";
import ModalLayout from "@/components/LayoutModal";
import styles from "@/styles/LayoutModal.module.css";
import React from "react";
import Image from "next/image";
import Input from "@/components/Input";

export default function RegisterModal() {
    const { activeModal, closeModal, } = useModal();

    if (activeModal !== "register")
        return null;

    return (
        <ModalLayout onClose={closeModal} defaultLayout={true}>
            <div className={styles.header}>
                <div className={styles.modalTitle}>Join Hypertube</div>
                <button className="clean-btn" onClick={closeModal}>
                    <Image className="black-icon" src="/icons/cross.svg" alt="cross" width={30} height={30}/>
                </button>
            </div>
            <Input type="email" placeholder="Email"></Input>
            {/*<input className="field" type="email" placeholder="Email"/>*/}
            <div className="flex">
                <Input type="firstname" placeholder="Firstname"></Input>
                <Input type="lastname" placeholder="Lastname"></Input>

                {/*<input className="field" type="firstname" placeholder="Firstname"/>*/}
                {/*<input className="field" type="lastname" placeholder="Lastname"/>*/}
            </div>

            <Input type="username" placeholder="Username" className={"max-70"}></Input>
            <Input type="password" placeholder="Password" className={"max-70"}></Input>

            {/*<input className="field max-70" type="username" placeholder="Username"/>*/}
            {/*<input className="field max-70" type="password" placeholder="Password"/>*/}

            <button className={styles.signInBtn + " action-btn max-40"}>Sign Up</button>
        </ModalLayout>
    );
}
