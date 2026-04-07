import React from "react";
import styles from "@/styles/Input.module.css";

export default function Input({type, placeholder, className }: {type: string, placeholder: string, className?: string}) {
    const isPassword = type === "password";

    return (<div className={styles.container}>
        <label htmlFor={type} className={styles.label}>{placeholder}</label>
        {isPassword ?? <a href="#">Forgotten?</a>}
        <input id={type} type={type} placeholder="" className={styles.field + " " + className}/>
    </div>);
}
