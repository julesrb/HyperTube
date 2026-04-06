"use client";

import styles from "@/styles/LayoutModal.module.css";
import React from "react";

export default function ModalLayout({ children, onClose, defaultLayout }: { children: React.ReactNode; onClose: () => void; defaultLayout: boolean; }) {
    if (!defaultLayout)
        return (<div onClick={onClose} className={styles.customBg} >{children}</div>);
    return (
        <div onClick={onClose} className={styles.defaultBg} >
            <div onClick={(e) => e.stopPropagation()} className={styles.modal + " border"}>
                <div className={styles.content}>
                    {children}
                </div>
            </div>
        </div>
    );
}
