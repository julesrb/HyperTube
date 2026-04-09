import React from "react";

export default function Input({type, placeholder, className }: {type: string, placeholder: string, className?: string}) {
    const isPassword = type === "password";

    return (<div className="flex flex-col w-full h-16">
        {isPassword ?? <a href="#">Forgotten?</a>}
        <input id={type} type={type} placeholder=""
               className={"peer py-6 m-0 w-full h-8 bg-white text-black border-b focus:border-b-2 "  + className}/>
        <label htmlFor={type}
               className="uppercase relative text-xs font-light bottom-14
               peer-focus:text-xs peer-focus:font-light peer-focus:bottom-14
               peer-placeholder-shown:font-normal peer-placeholder-shown:font-bold peer-placeholder-shown:bottom-9 peer-placeholder-shown:text-xl">{placeholder}</label>
    </div>);
}
