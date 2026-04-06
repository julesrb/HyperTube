import styles from "@/styles/LanguageDropdown.module.css";
import React, {useState} from "react";

export default function LanguageDropdown(Icon: React.JSX.Element, className: string) {
    const [isOpen, setIsOpen] = useState(false);
    const [selectedLanguage, setSelectedLanguage] = useState([true, false, false]);
    const languages = ["English", "French", "Spanish"];

    const handleSwitchLanguage = (index: number) => {
        const newLanguage = [false, false, false];
        newLanguage[index] = true;
        setSelectedLanguage(newLanguage);
        setIsOpen(false);
    }

    return (
        <div
            className={className}
            onMouseEnter={() => (setIsOpen(true))}
            onMouseLeave={() => (setIsOpen(false))}>

            {Icon}
            {isOpen && <div className={styles.dropdown + " border"}>
                {languages.map((language, i) => (
                    <button key={language} className={"underline clean-btn hairline" + (selectedLanguage[i] ? " selected" : "")} onClick={() => handleSwitchLanguage(i)}>{language}</button>
                ))}
            </div>}
        </div>
    );
}
