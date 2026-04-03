"use client";

import styles from "@/styles/LayoutModal.module.css";
import React from "react";
import Image from "next/image";

export default function ModalLayout({ children, onClose, }: { children: React.ReactNode; onClose: () => void; }) {
    return (
        <div onClick={onClose} className={styles.bg} >
            <div onClick={(e) => e.stopPropagation()} className={styles.modal + " border"}>
                <div className={styles.header}>
                    <button className="clean-btn" onClick={onClose}>
                        <Image className="black-cross" src="/icons/cross.svg" alt="cross" width={30} height={30}/>
                    </button>
                </div>
                <div className={styles.content}>
                    {children}
                </div>
            </div>
        </div>
    );
}
