"use client";

import { useModal } from "@/context/ModalContext";
import ModalLayout from "@/components/LayoutModal";
import styles from "@/styles/LayoutModal.module.css";
import Image from "next/image";
import React from "react";
import Input from "@/components/Input";

export default function SigninModal() {
    const {activeModal, closeModal,} = useModal();

    if (activeModal !== "signin")
        return null;

    return (
        <ModalLayout onClose={closeModal} defaultLayout={false}>
            <div className={styles.signInModal + " border"} onClick={(e) => e.stopPropagation()}>
                <button className="clean-btn" onClick={closeModal}>
                    <Image className="black-icon" src="/icons/cross.svg" alt="cross" width={30} height={30}/>
                </button>
                <Input type="email" placeholder="Email"></Input>
                <Input type="password" placeholder="Password"></Input>
                {/*<input type="email" placeholder="Email" className="field border"/>*/}
                {/*<input type="password" placeholder="password" className="field border"/>*/}

                <div className="flex" style={{alignItems: "flex-end", marginRight: "8px"}}>
                    <button className={styles.signInBtn + " action-btn"}>Sign In</button>
                </div>
            </div>
        </ModalLayout>
    );
}