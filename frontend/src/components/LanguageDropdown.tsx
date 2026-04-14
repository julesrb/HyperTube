import React, {useState} from "react";

export default function LanguageDropdown(Icon: React.JSX.Element) {
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
            className="flex items-center hover:cursor-pointer"
            onMouseEnter={() => (setIsOpen(true))}
            onMouseLeave={() => (setIsOpen(false))}>

            {Icon}
            {isOpen && <div className="absolute top-10 right-10 z-50 p-8">
                <div className="flex flex-col gap-1 items-start bg-white py-4 px-5 custom-shadow-s border">
                    {languages.map((language, i) => (
                        <button key={language} className={"custom-h-underline text-xl font-hairline" + (selectedLanguage[i] ? " custom-selected" : "")} onClick={() => handleSwitchLanguage(i)}>{language}</button>
                    ))}
                </div>
            </div>}
        </div>
    );
}
