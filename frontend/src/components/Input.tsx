import React from "react";
import {EyeIcon} from "@/components/Icon";

export default function Input({id, type, placeholder, value, onChange, className }: {id: string, type: string, placeholder: string, value: string, onChange: (e: React.ChangeEvent<HTMLInputElement>, key: string) => void, className?: string}) {
    const isPassword = type === "password";
    const [isPasswordVisible, setIsPasswordVisible] = React.useState(false);

    const handleTogglePasswordVisibility = () => {
        setIsPasswordVisible(!isPasswordVisible);
    }

    return (<div className={"flex flex-col w-full h-16 relative " + className}>
        <input id={id} type={isPasswordVisible && isPassword ? "text" : type} placeholder=""
               value={value}
               onChange={(e) => onChange(e, id)}
               className="peer py-4 m-0 w-full h-8 bg-white text-black border-b focus:border-b-2"/>
        <label htmlFor={type}
               className="pointer-events-none uppercase absolute text-xs font-sans bottom-15
               peer-focus:text-xs peer-focus:font-sans peer-focus:bottom-15
               peer-placeholder-shown:font-condensed peer-placeholder-shown:tracking-wide peer-placeholder-shown:bottom-9 peer-placeholder-shown:text-2xl">{placeholder}</label>
        {isPassword && (<button className="absolute right-0 top-1" onClick={handleTogglePasswordVisibility}><EyeIcon crossed={isPasswordVisible}/></button>)}
    </div>
    );
}
