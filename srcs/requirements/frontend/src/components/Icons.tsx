"use client";

import {useEffect, useState} from "react";
import {tOauthService} from "@/types/user";

export function CrossIcon({color="black", size=30, className=""}) {
    const fullColor = `var(--color-${color})`;

    return (<svg className={className} width={size} height={size} viewBox="0 0 36 36" fill="none" xmlns="http://www.w3.org/2000/svg">
        <line x1="9.32258" y1="8.613" x2="27.2582" y2="26.5486" stroke={fullColor}/>
        <line x1="27.2584" y1="9.32279" x2="9.32279" y2="27.2584" stroke={fullColor}/>
    </svg>);
}

export function RightIcon({color="black", size=13}) {
    const fullColor = `var(--color-${color})`;

    return (<svg width={size} height={size} viewBox="0 0 16 27" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path d="M1.97437 24.8514L12.4744 13.3514L1.47437 1.35144" stroke={fullColor} strokeWidth="4"/>
    </svg>);
}

export function LeftIcon({color="black", size=13}) {
    const fullColor = `var(--color-${color})`;

    return (<svg width={size} height={size} viewBox="0 0 16 27" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path d="M13.2107 1.34854L2.71069 12.8485L13.7107 24.8485" stroke={fullColor} strokeWidth="4"/>
    </svg>);
}

export function UserIcon({selected, color="black", size=20}: {selected: boolean, color?: string, size?: number}) {
    const fullColor = `var(--color-${color})`;

    if (selected) {
        return (<svg width={size} height={size} viewBox="0 0 18 23" fill="none" xmlns="http://www.w3.org/2000/svg">
            <circle cx="9" cy="5" r="4.5" fill={fullColor} stroke={fullColor}/>
            <path d="M0.511719 20.5879L0.510742 20.5752C0.268441 15.6859 4.19964 11.5 9 11.5C13.7908 11.5 17.7145 15.6692 17.4893 20.5459C17.4567 20.7138 17.2975 20.9506 16.835 21.2158C16.3717 21.4814 15.7019 21.7182 14.8701 21.9131C13.2124 22.3015 11.0205 22.5 8.82617 22.5C6.63112 22.5 4.47028 22.3014 2.87305 21.9141C2.07047 21.7194 1.44426 21.4844 1.03027 21.2256C0.605388 20.9599 0.511719 20.7418 0.511719 20.5996V20.5879Z" fill={fullColor} stroke={fullColor}/>
        </svg>);
    }
    return (<svg width={size} height={size} viewBox="0 0 18 23" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path d="M8.99951 11C14.0932 11.0003 18.2436 15.433 17.9878 20.5996C17.5322 23.7494 0.556311 23.7991 0.0239258 20.748L0.0112305 20.5996C-0.244646 15.4328 3.90558 11 8.99951 11ZM8.99951 12C4.49281 12 0.780709 15.9391 1.00928 20.5508L1.01025 20.5635C1.03025 20.5957 1.0971 20.6784 1.29443 20.8018C1.64261 21.0194 2.21027 21.2384 2.99072 21.4277C4.53596 21.8024 6.65404 22 8.82568 22C10.9962 22 13.1481 21.8036 14.7563 21.4268C15.566 21.237 16.1823 21.0134 16.5854 20.7822C16.8782 20.6143 16.9671 20.496 16.9917 20.4561C17.1654 15.8853 13.4752 12.0003 8.99951 12ZM8.99951 0C11.7607 0.000263611 13.9995 2.23874 13.9995 5C13.9995 7.76126 11.7607 9.99974 8.99951 10C6.23809 10 3.99951 7.76142 3.99951 5C3.99951 2.23858 6.23809 0 8.99951 0ZM8.99951 1C6.79037 1 4.99951 2.79086 4.99951 5C4.99951 7.20914 6.79037 9 8.99951 9C11.2084 8.99974 12.9995 7.20898 12.9995 5C12.9995 2.79102 11.2084 1.00026 8.99951 1Z" fill={fullColor}/>
    </svg>);
}

export function SearchIcon({selected, color="black", size=20 }: {selected: boolean, color?: string, size?: number}) {
    const fullColor = `var(--color-${color})`;

    if (selected) {
        return (<svg width={size} height={size} viewBox="0 0 27 27" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M9.5 0C14.7467 0 19 4.25329 19 9.5C19 11.5618 18.3404 13.4682 17.2246 15.0254L27 24.2119L24.9453 26.3984L15.1152 17.1602C13.542 18.3154 11.6015 19 9.5 19C4.25329 19 0 14.7467 0 9.5C0 4.25329 4.25329 0 9.5 0ZM9.5 3C5.91015 3 3 5.91015 3 9.5C3 13.0899 5.91015 16 9.5 16C13.0899 16 16 13.0899 16 9.5C16 5.91015 13.0899 3 9.5 3Z" fill={fullColor}/>
        </svg>);
    }
    return (<svg width={size} height={size} viewBox="0 0 26 26" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path d="M9.5 0C14.7467 0 19 4.25329 19 9.5C19 12.0712 17.9766 14.4019 16.3174 16.1123L26 25.2119L25.3154 25.9404L15.584 16.7949C13.9358 18.171 11.815 19 9.5 19C4.25329 19 0 14.7467 0 9.5C0 4.25329 4.25329 0 9.5 0ZM9.5 1C4.80558 1 1 4.80558 1 9.5C1 14.1944 4.80558 18 9.5 18C14.1944 18 18 14.1944 18 9.5C18 4.80558 14.1944 1 9.5 1Z" fill={fullColor}/>
    </svg>);
}

export function RegisterIcon({color="black", size=20}) {
    const fullColor = `var(--color-${color})`;

    return (<svg width={size} height={size} viewBox="0 0 21 23" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path d="M16.731 18.0498H20.7808V18.9502H16.731V23H15.8306V18.9502H11.7808V18.0498H15.8306V14H16.731V18.0498ZM8.99951 11C10.3007 11.0001 11.5403 11.2906 12.6606 11.8076C12.2737 11.99 11.9026 12.2011 11.5503 12.4375C10.7471 12.1557 9.88893 12.0001 8.99951 12C4.49281 12 0.780709 15.9391 1.00928 20.5508L1.01025 20.5635C1.03042 20.5958 1.09754 20.6786 1.29443 20.8018C1.64264 21.0193 2.21053 21.2385 2.99072 21.4277C4.37577 21.7635 6.22142 21.9557 8.15283 21.9922C8.2584 22.3369 8.38701 22.6716 8.53369 22.9961C4.33845 22.9629 0.280074 22.2151 0.0239258 20.748L0.0112305 20.5996C-0.244646 15.4328 3.90558 11 8.99951 11ZM8.99951 0C11.7605 0.000527208 13.9995 2.2389 13.9995 5C13.9995 7.7611 11.7605 9.99947 8.99951 10C6.23809 10 3.99951 7.76142 3.99951 5C3.99951 2.23858 6.23809 4.81392e-07 8.99951 0ZM8.99951 1C6.79037 1 4.99951 2.79086 4.99951 5C4.99951 7.20914 6.79037 9 8.99951 9C11.2082 8.99947 12.9995 7.20881 12.9995 5C12.9995 2.79119 11.2082 1.00053 8.99951 1Z" fill={fullColor}/>
    </svg>);
}

export function LanguageIcon({selected,color="black", size=24}: {selected: boolean, color?: string, size?: number}) {
    const fullColor = `var(--color-${color})`;
    const fullWhite = `var(--color-white)`;

    if (selected) {
        return (<svg width={size} height={size} viewBox="0 0 27 27" fill="none" xmlns="http://www.w3.org/2000/svg">
            <circle cx="13.5" cy="13.5" r="13" fill={fullColor} stroke={fullColor}/>
            <circle cx="13.3262" cy="13.5" r="11" stroke={fullWhite}/>
            <path d="M13.3262 2.5C14.5418 2.5 15.7906 3.54737 16.7646 5.58398C17.7211 7.58399 18.3262 10.3822 18.3262 13.5C18.3262 16.6178 17.7211 19.416 16.7646 21.416C15.7906 23.4526 14.5418 24.5 13.3262 24.5C12.1105 24.5 10.8617 23.4526 9.8877 21.416C8.93121 19.416 8.32617 16.6178 8.32617 13.5C8.32617 10.3822 8.93121 7.58399 9.8877 5.58398C10.8617 3.54737 12.1105 2.5 13.3262 2.5Z" stroke={fullWhite}/>
            <line x1="2.82617" y1="9.5" x2="24" y2="9.5" stroke={fullWhite}/>
            <line x1="2.82617" y1="15.5" x2="24.8262" y2="15.5" stroke={fullWhite}/>
        </svg>);
    }
    return (<svg width={size} height={size} viewBox="0 0 27 27" fill="none" xmlns="http://www.w3.org/2000/svg">
        <circle cx="13.5" cy="13.5" r="13" fill={fullWhite} stroke={fullWhite}/>
        <circle cx="13.3262" cy="13.5" r="11" stroke={fullColor}/>
        <path d="M13.3262 2.5C14.5418 2.5 15.7906 3.54737 16.7646 5.58398C17.7211 7.58399 18.3262 10.3822 18.3262 13.5C18.3262 16.6178 17.7211 19.416 16.7646 21.416C15.7906 23.4526 14.5418 24.5 13.3262 24.5C12.1105 24.5 10.8617 23.4526 9.8877 21.416C8.93121 19.416 8.32617 16.6178 8.32617 13.5C8.32617 10.3822 8.93121 7.58399 9.8877 5.58398C10.8617 3.54737 12.1105 2.5 13.3262 2.5Z" stroke={fullColor}/>
        <line x1="2.82617" y1="9.5" x2="24" y2="9.5" stroke={fullColor}/>
        <line x1="2.82617" y1="15.5" x2="24.8262" y2="15.5" stroke={fullColor}/>
    </svg>);
}

export function EyeIcon({color="black", size=24, crossed=false}: {color?: string, size?: number, crossed?: boolean}) {
    const fullColor = `var(--color-${color})`;

    if (crossed)
        return (<svg height={size} width={size} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" className="size-5">
            <g clipPath="url(#clip0_2902_56205)">
                <path
                    d="M12.0001 4.0645C9.61239 4.0645 7.67013 4.77712 6.12007 5.77699L18.4148 17.9851C21.6687 15.6147 23.0001 12.0645 23.0001 12.0645C23.0001 12.0645 20.0001 4.0645 12.0001 4.0645ZM12.0001 7.0645C14.7611 7.0645 17.0001 9.3035 17.0001 12.0645C17.0001 13.3004 16.5514 14.4318 15.8081 15.3045L8.64004 8.36201C9.52782 7.55584 10.7067 7.0645 12.0001 7.0645ZM12.0001 9.0645C11.5336 9.0645 11.0784 9.17312 10.6682 9.37635L14.6875 13.3979C14.8912 12.9873 15.0001 12.5315 15.0001 12.0645C15.0001 11.2688 14.684 10.5058 14.1214 9.94318C13.5588 9.38057 12.7957 9.0645 12.0001 9.0645Z"
                    fill={fullColor} />
                <path fillRule="evenodd" clipRule="evenodd"
                      d="M22.343 23.7651L0.28125 1.70335L1.6389 0.345703L23.7006 22.4074L22.343 23.7651Z"
                      fill={fullColor} />
                <path d="M11.8195 15.059C10.305 14.969 9.09286 13.7556 9.00513 12.2405L11.8195 15.059Z"
                      fill={fullColor} />
                <path
                    d="M6.99561 12.0645C6.99561 14.7894 9.21646 17.0688 12 17.0688C12.5652 17.0688 13.0909 16.9785 13.5982 16.8068L16.0873 19.2959C12.6987 20.6758 5.05674 21.1038 0.995605 12.0645C1.37115 11.13 2.38425 9.12997 4.13971 7.37451L7.25761 10.4924C6.99561 11.1736 6.99561 12.0645 6.99561 12.0645Z"
                    fill={fullColor} />
            </g>
            <defs>
                <clipPath id="clip0_2902_56205">
                    <rect width="24" height="24" fill={fullColor} />
                </clipPath>
            </defs>
        </svg>);

    return (<svg height={size} width={size} viewBox="0 0 24 25" fill="none" xmlns="http://www.w3.org/2000/svg" className="size-5">
        <path
            d="M12 4.08398C4 4.08398 1 12.084 1 12.084C1 12.084 4 20.084 12 20.084C20 20.084 23 12.084 23 12.084C23 12.084 20 4.08398 12 4.08398ZM12 7.08398C14.761 7.08398 17 9.32298 17 12.084C17 14.845 14.761 17.084 12 17.084C9.239 17.084 7 14.845 7 12.084C7 9.32298 9.239 7.08398 12 7.08398ZM12 9.08398C11.2044 9.08398 10.4413 9.40005 9.87868 9.96266C9.31607 10.5253 9 11.2883 9 12.084C9 12.8796 9.31607 13.6427 9.87868 14.2053C10.4413 14.7679 11.2044 15.084 12 15.084C12.7956 15.084 13.5587 14.7679 14.1213 14.2053C14.6839 13.6427 15 12.8796 15 12.084C15 11.2883 14.6839 10.5253 14.1213 9.96266C13.5587 9.40005 12.7956 9.08398 12 9.08398Z"
            fill={fullColor} />
    </svg>);
}

export function HypertubResponsiveLogo({color="black", height=20}: {color?: string, height?: number, width?: number, className?: string}) {
    const fullColor = `var(--color-${color})`;
    const defaultRecWidth = 266;
    const minWindowWidth = 640;
    const [recWidth, setRecWidth] = useState(defaultRecWidth);

    const handleResize = (value: number) => value - defaultRecWidth + recWidth;

    useEffect(() => {
        const handlWindowResize = () => {
            let newRecWidth = (window.innerWidth - minWindowWidth) / 3;
            if (newRecWidth < 0)
                newRecWidth = 0;
            setRecWidth(newRecWidth);
        };
        handlWindowResize();
        window.addEventListener("resize", handlWindowResize);

        return () => window.removeEventListener("resize", handlWindowResize);
    }, []);

    return (<svg className="h-3 sm:h-6 md:h-4 xl:h-6 w-auto" width={height * handleResize(1110) / 65} height={height} viewBox={`0 0 ${handleResize(1110)} 65`} fill="none" xmlns="http://www.w3.org/2000/svg">
        {/*h*/} <path d="M590.15 30.5918C590.15 41.8558 595.142 48.1279 610.758 48.1279C611.645 48.1279 612.498 48.1057 613.317 48.0654V64.6094C612.48 64.6282 611.626 64.6396 610.758 64.6396C579.014 64.6396 567.75 52.0958 567.75 30.5918V0H590.15V30.5918Z" fill={fullColor}/>
        {/*y*/} <path d="M22.5283 23.4238H63.7441V0H86.2725V64H63.7441V40.0645H22.5283V64H0V0H22.5283V23.4238Z" fill={fullColor}/>
        {/*p*/} <path d="M143.669 25.0879L166.325 0H194.613L154.805 39.2959V64H132.277V39.2959L92.4688 0H120.757L143.669 25.0879Z" fill={fullColor}/>
        {/*e*/} <path fillRule="evenodd" clipRule="evenodd" d="M253.483 0C273.451 7.96847e-05 285.099 6.0165 285.099 20.6084C285.098 35.2001 273.451 41.2158 252.971 41.2158H223.403V64H200.875V0H253.483ZM223.403 25.4717H252.587C258.987 25.4717 262.699 24.4481 262.699 20.6084C262.699 16.7684 258.987 15.8721 252.587 15.8721H223.403V25.4717Z" fill={fullColor}/>
        {/*r*/} <path d="M374.25 15.8721H316.778V23.9355H371.946V39.8076H316.778V48.1279H374.25V64H294.25V0H374.25V15.8721Z" fill={fullColor}/>
        {/*t*/} <path fillRule="evenodd" clipRule="evenodd" d="M440.483 0C460.451 5.76439e-05 472.098 4.35202 472.099 18.4316C472.099 28.6716 463.907 32.8956 454.179 34.5596C463.395 35.3276 467.747 38.2719 469.411 46.4639L471.331 55.9355C472.099 59.7752 472.483 61.5677 473.635 62.8477V64H451.363C449.571 61.8241 449.315 59.2642 448.547 56.3203L446.499 48.1279C445.219 42.6241 443.043 40.3204 434.979 40.3203H410.403V64H387.875V0H440.483ZM410.403 26.2402H438.819C445.987 26.2402 449.699 25.216 449.699 20.9922C449.699 16.6403 445.987 15.8721 438.819 15.8721H410.403V26.2402Z" fill={fullColor}/>
        {/*u 1/2*/} <path d="M558.469 16.6396H530.181V64H507.909V16.6396H479.621V0H558.469V16.6396Z" fill={fullColor}/>
        {/*u 2/2*/} <path d={`M${handleResize(917.01)} 30.5918C${handleResize(917.01)} 52.0958 ${handleResize(905.874)} 64.6396 ${handleResize(874.258)} 64.6396C${handleResize(873.602)} 64.6396 ${handleResize(872.956)} 64.6318 ${handleResize(872.317)} 64.6211V48.0928C872.945 48.1155 ${handleResize(873.592)} 48.1279 ${handleResize(874.258)} 48.1279C${handleResize(889.874)} 48.1279 ${handleResize(894.866)} 41.8558 ${handleResize(894.866)} 30.5918V0H${handleResize(917.01)}V30.5918Z`} fill={fullColor}/>
        {/*b*/} <path fillRule="evenodd" clipRule="evenodd" d={`M${handleResize(982.98)} 0C${handleResize(1001.54)} 6.95936e-05 ${handleResize(1013.96)} 4.09606 ${handleResize(1013.96)} 16.3838C${handleResize(1013.96)} 23.4238 ${handleResize(1008.71)} 27.2637 ${handleResize(1000.26)} 29.0557C${handleResize(1010.24)} 30.7196 ${handleResize(1016.64)} 34.9439 ${handleResize(1016.64)} 44.9277C${handleResize(1016.64)} 58.6237 ${handleResize(1003.33)} 64 ${handleResize(984.516)} 64H${handleResize(930.5)}V0H${handleResize(982.98)}ZM${handleResize(952.9)} 48.1279H${handleResize(984.26)}C${handleResize(990.532)} 48.1279 ${handleResize(994.244)} 47.3595 ${handleResize(994.244)} 43.5195C${handleResize(994.244)} 39.5519 ${handleResize(990.532)} 38.7842 ${handleResize(984.26)} 38.7842H${handleResize(952.9)}V48.1279ZM${handleResize(952.9)} 24.5762H${handleResize(983.492)}C${handleResize(989.38)} 24.5762 ${handleResize(993.22)} 23.9357 ${handleResize(993.22)} 20.0957C${handleResize(993.22)} 16.512 ${handleResize(989.38)} 15.8721 ${handleResize(983.492)} 15.8721H${handleResize(952.9)}V24.5762Z`} fill={fullColor}/>
        {/*e*/} <path d={`M${handleResize(1109.12)} 15.8721H${handleResize(1051.65)}V23.9355H${handleResize(1106.82)}V39.8076H${handleResize(1051.65)}V48.1279H${handleResize(1109.12)}V64H${handleResize(1029.12)}V0H${handleResize(1109.12)}V15.8721Z`} fill={fullColor}/>
        <rect x="608.317" y="48" width={recWidth} height="16.5" fill={fullColor}/>
    </svg>);
}

export function HypertubeLogo({color="black", height=20, width, className}: {color?: string, height?: number, width?: number, className?: string}) {
    const fullColor = `var(--color-${color})`;
    const preserveRatio = height && width ? "none" : "xMidYMid meet";
    if (!width)
        width = (height * 347) / 21
    if (!height)
        height = (width * 21) / 347

    return (<svg className={className} width={width} height={height} viewBox="0 0 347 21" fill="none" xmlns="http://www.w3.org/2000/svg" preserveAspectRatio={preserveRatio}>
        <path d="M7.04004 7.31934H19.9199V0H26.96V20H19.9199V12.5195H7.04004V20H0V0H7.04004V7.31934Z" fill={fullColor}/>
        <path d="M44.8965 7.83984L51.9766 0H60.8164L48.377 12.2793V20H41.3369V12.2793L28.8965 0H37.7363L44.8965 7.83984Z" fill={fullColor}/>
        <path fillRule="evenodd" clipRule="evenodd" d="M79.2139 0C85.4536 8.58141e-05 89.0937 1.87964 89.0938 6.43945C89.0938 10.9994 85.4535 12.8798 79.0537 12.8799H69.8135V20H62.7734V0H79.2139ZM69.8135 7.95996H78.9336C80.9335 7.95994 82.0938 7.63942 82.0938 6.43945C82.0937 5.23958 80.9334 4.95998 78.9336 4.95996H69.8135V7.95996Z" fill={fullColor}/>
        <path d="M116.953 4.95996H98.9932V7.47949H116.233V12.4395H98.9932V15.04H116.953V20H91.9531V0H116.953V4.95996Z" fill={fullColor}/>
        <path fillRule="evenodd" clipRule="evenodd" d="M137.651 0C143.891 6.20791e-05 147.531 1.35987 147.531 5.75977C147.531 8.95963 144.971 10.2798 141.931 10.7998C144.811 11.0398 146.171 11.9595 146.691 14.5195L147.291 17.4795C147.531 18.6794 147.651 19.2397 148.011 19.6396V20H141.051C140.491 19.32 140.411 18.5196 140.171 17.5996L139.531 15.04C139.131 13.32 138.451 12.5996 135.931 12.5996H128.251V20H121.211V0H137.651ZM128.251 8.19922H137.131C139.371 8.19922 140.531 7.87957 140.531 6.55957C140.531 5.19957 139.371 4.95996 137.131 4.95996H128.251V8.19922Z" fill={fullColor}/>
        <path d="M174.521 5.19922H165.682V20H158.722V5.19922H149.882V0H174.521V5.19922Z" fill={fullColor}/>
        <path fillRule="evenodd" clipRule="evenodd" d="M306.618 0C312.418 2.13266e-05 316.298 1.27951 316.298 5.11914C316.298 7.31914 314.658 8.5191 312.018 9.0791C315.138 9.5991 317.138 10.92 317.138 14.04C317.137 18.3197 312.977 20 307.098 20H290.218V0H306.618ZM297.218 15.04H307.018C308.978 15.04 310.138 14.7996 310.138 13.5996C310.138 12.3596 308.978 12.1191 307.018 12.1191H297.218V15.04ZM297.218 7.67969H306.778C308.618 7.67967 309.818 7.47923 309.818 6.2793C309.818 5.1596 308.618 4.95998 306.778 4.95996H297.218V7.67969Z" fill={fullColor}/>
        <path d="M346.038 4.95996H328.078V7.47949H345.318V12.4395H328.078V15.04H346.038V20H321.038V0H346.038V4.95996Z" fill={fullColor}/>
        <path fillRule="evenodd" clipRule="evenodd" d="M190.862 15.04C185.982 15.04 184.422 13.0796 184.422 9.55957V0H177.422V9.55957C177.422 16.2796 180.942 20.1992 190.862 20.1992H272.643C282.522 20.1992 286.002 16.2795 286.002 9.55957V0H279.082V9.55957C279.082 13.0107 277.583 14.9611 272.925 15.0361L190.862 15.04Z" fill={fullColor}/>
    </svg>);
}

export function ExitDoorIcon({selected, color="black", size=20}: {selected: boolean, color?: string, size?: number}) {
    const fullColor = `var(--color-${color})`;

    if (selected) {
        return (<svg width={size} height={size} viewBox="0 0 22 23" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M20.75 0C21.3023 1.93276e-07 21.75 0.447715 21.75 1V22C21.75 22.5523 21.3023 23 20.75 23H8.75C8.19772 23 7.75 22.5523 7.75 22V14H12.8311C13.1072 14 13.3311 13.7761 13.3311 13.5V10.5C13.3311 10.2239 13.1072 10 12.8311 10H7.75V1C7.75 0.447715 8.19772 2.41596e-08 8.75 0H20.75Z" fill={fullColor}/>
            <path d="M0.146446 11.6464C-0.0488157 11.8417 -0.0488157 12.1583 0.146446 12.3536L3.32843 15.5355C3.52369 15.7308 3.84027 15.7308 4.03553 15.5355C4.2308 15.3403 4.2308 15.0237 4.03553 14.8284L1.20711 12L4.03553 9.17157C4.2308 8.97631 4.2308 8.65973 4.03553 8.46447C3.84027 8.2692 3.52369 8.2692 3.32843 8.46447L0.146446 11.6464ZM12 12V11.5L0.5 11.5V12V12.5L12 12.5V12Z" fill={fullColor}/>
        </svg>);
    }
    return (<svg width={size} height={size} viewBox="0 0 23 24" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path d="M20.75 0C21.5784 2.06158e-05 22.25 0.671585 22.25 1.5V22.5C22.25 23.3284 21.5784 24 20.75 24H8.75C7.92157 24 7.25 23.3284 7.25 22.5V18.5H8.25V22.5C8.25 22.7761 8.47386 23 8.75 23H20.75C21.0261 23 21.25 22.7761 21.25 22.5V1.5C21.25 1.22387 21.0261 1.00002 20.75 1H8.75C8.47386 1 8.25 1.22386 8.25 1.5V6.5H7.25V1.5C7.25 0.671573 7.92157 0 8.75 0H20.75ZM3.32812 8.96484C3.52339 8.7696 3.8399 8.76959 4.03516 8.96484C4.23039 9.1601 4.23039 9.47662 4.03516 9.67188L1.70703 12H12V13H1.70703L4.03516 15.3281C4.23039 15.5234 4.23039 15.8399 4.03516 16.0352C3.8399 16.2304 3.52339 16.2304 3.32812 16.0352L0.146484 12.8535C-0.0487776 12.6583 -0.0487776 12.3417 0.146484 12.1465L3.32812 8.96484Z" fill={fullColor}/>
    </svg>);
}

export function CheckIcon({color="black", size=20, className="" }) {
    const fullColor = `var(--color-${color})`;

    return (<svg className={className} width={size} height={size} viewBox="0 0 18 15" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path d="M16.7087 0.401886L6.653 13.184L0.459512 6.99635" stroke={fullColor} strokeWidth="1.1"/>
    </svg>);
}

export function GridIcon({color="black", size=20}) {
    const fullColor = `var(--color-${color})`;

    return (<svg height={size} width={size} fill="none" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 22 15">
        <path fill={fullColor} stroke={fullColor} strokeWidth="1.5" d="M1.252 1.697h4.5v4.5h-4.5zM8.752 1.697h4.5v4.5h-4.5zM16.252 1.697h4.5v4.5h-4.5zM1.252 9.197h4.5v4.5h-4.5zM8.752 9.197h4.5v4.5h-4.5zM16.252 9.197h4.5v4.5h-4.5z" />
    </svg>);
}

export function ListIcon({color="black", size=20}) {
    const fullColor = `var(--color-${color})`;

    return (<svg height={size} width={size} fill="none" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 22 14">
        <path fillRule="evenodd" clipRule="evenodd" d="M21.499 3.33H5.719V.505h15.78V3.33ZM3.325 3.33H.5V.505h2.825V3.33ZM21.499 8.649H5.719V5.824h15.78v2.825ZM3.325 8.661H.5V5.836h2.825v2.825ZM21.499 13.966H5.719v-2.825h15.78v2.825ZM3.325 13.964H.5V11.14h2.825v2.825Z"
              fill={fullColor} />
    </svg>);
}

export function SortIcon({sideUp=true, color="black", size=15}) {
    const fullColor = `var(--color-${color})`;

    return (<svg height={size} width={size} fill="none" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 16">
        {sideUp ?
            <path d="M5.95 4.68171e-06L4.13909 4.99834e-06L4.13909 11.74L1.30182 9.11728L-8.4212e-07 10.3827L5.04455 15.1991L9.62091 10.3827L8.32 9.11637L5.95 11.7391L5.95 4.68171e-06Z" fill={fullColor} /> :
            <path d="M5.95 15.1991L4.13909 15.1991L4.13909 3.45909L1.30182 6.08182L-8.4212e-07 4.81636L5.04455 8.82016e-07L9.62091 4.81637L8.32 6.08273L5.95 3.46L5.95 15.1991Z" fill={fullColor} />}
    </svg>);
}

export function StarIcon({color="yellow", size=20}) {
    const fullColor = `var(--color-${color})`;

    return (<svg height={size} width={size} fill="none" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 22 20">
        <path d="M10.4617 0L12.9313 7.60081H20.9233L14.4577 12.2984L16.9273 19.8992L10.4617 15.2016L3.99603 19.8992L6.46569 12.2984L4.86374e-05 7.60081H7.99202L10.4617 0Z"
            fill={fullColor}/>
    </svg>);
}

export function EditIcon({color="black", size=20}) {
    const fullColor = `var(--color-${color})`;

    return (<svg width={size} height={size} xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512">
        <path fill={fullColor} d="M375.27,512H49.9c-26.35-7.32-45.43-27.75-49.9-54.89V142.72c6.12-32.56,30.14-53.9,63.33-55.94,55.32-3.39,114.07,2.67,169.75.01,26.7.78,30.92,36.54,4.71,42.71l-176.42.24c-8.55.96-14.91,7.3-17.99,14.95l-.03,310.48c1.52,7.16,10.42,14.94,17.98,14.96h302.49c10.43-.77,18.58-9.16,19.41-19.51l.12-174.54c4.86-26.51,41.1-23.67,42.79,2.9-2.45,55.73,3.32,114.31.06,169.7-1.86,31.63-20.1,55.28-50.95,63.33Z"/>
        <path fill={fullColor} d="M367.92,69.94l74.78,75.77-168.3,168.11c-25.57,8.51-53.13,11.12-79.4,17.42-8.68,1.48-14.72-4.52-13.23-13.23,6.3-26.28,8.91-53.83,17.42-79.4l168.74-168.66Z"/>
        <path fill={fullColor} d="M451.91,2.29c33.16-3.52,61.48,23.82,59.07,57.08-2.02,27.95-27.2,43.57-44.53,62.32l-75.14-74.42c-.23-1.21.82-1.64,1.39-2.39,5.72-7.45,27.47-28.58,35.15-33.72,6.99-4.68,15.64-7.98,24.06-8.87Z"/>
    </svg>);
}

export function TrashIcon({color="red", size=20}) {
    const fullColor = `var(--color-${color})`;

    return (<svg width={size} height={size} xmlns="http://www.w3.org/2000/svg" viewBox="0 0 384 384">
        <path fill={fullColor} d="M92.82,384c-6.92-2.18-12.43-7.22-13.39-14.71l-15.8-239.54h256l-17.28,241.06c-1.47,7.36-6.61,11.03-13.41,13.19H92.82ZM160.68,167.5c-5.38-5.38-14.87-3.04-17.77,3.82l-.71,156.82c2.41,13.03,19.04,12.92,21.74,0l-.59-156.19c-.89-1.29-1.62-3.39-2.68-4.45ZM229.54,164.52c-4.01.16-8.59,3.45-9.71,7.36l-.51,156.25c1.96,13.36,20.26,12.83,21.71-.75l-.51-154.74c-1.49-4.8-5.87-8.33-10.98-8.12Z"/>
        <path fill={fullColor} d="M247.02,0c2.48.69,8.23,5.37,8.23,7.88v28.88h-127.25V7.88c0-2.5,5.75-7.19,8.23-7.88h110.78Z"/>
        <path fill={fullColor} d="M345.08,108H38.18v-26.62c0-11.62,13-22.83,24.34-22.86l256.77-.05c11.93-.61,25.79,10.7,25.79,22.91v26.62Z"/>
    </svg>);
}

export function OAuthIcon({oauth, color="white", size=30}: {oauth: tOauthService, color?: string, size?: number}) {
    const fullColor = `var(--color-${color})`;

    if (oauth === "42") {
        return (<svg width={size} height={size} viewBox="0 0 67 43" fill="none" xmlns="http://www.w3.org/2000/svg" className="pt-0.5">
            <path d="M37.0271 0H24.6828L0 22.1433V31.1061H24.6828V42.1777H37.0271V22.1433H12.3385L37.0271 0Z" fill={fullColor}/>
            <path d="M42.3171 11.0716L54.6556 0H42.3171V11.0716Z" fill={fullColor}/>
            <path d="M66.9999 11.0716V0H54.6556V11.0716L42.3171 22.1433V33.2149H54.6556V22.1433L66.9999 11.0716Z" fill={fullColor}/>
            <path d="M66.999 22.1435L54.655 33.2152H66.999V22.1435Z" fill={fullColor}/>
        </svg>);
    } else if (oauth === "gitlab") {
        return (<svg width={size} height={size} className="tanuki-logo" viewBox="0 0 50 48" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path className="tanuki-shape tanuki" d="m49.014 19-.067-.18-6.784-17.696a1.792 1.792 0 0 0-3.389.182l-4.579 14.02H15.651l-4.58-14.02a1.795 1.795 0 0 0-3.388-.182l-6.78 17.7-.071.175A12.595 12.595 0 0 0 5.01 33.556l.026.02.057.044 10.32 7.734 5.12 3.87 3.11 2.351a2.102 2.102 0 0 0 2.535 0l3.11-2.352 5.12-3.869 10.394-7.779.029-.022a12.595 12.595 0 0 0 4.182-14.554Z" fill={fullColor}/>
            <path className="tanuki-shape right-cheek" d="m49.014 19-.067-.18a22.88 22.88 0 0 0-9.12 4.103L24.931 34.187l9.485 7.167 10.393-7.779.03-.022a12.595 12.595 0 0 0 4.175-14.554Z" fill="var(--color-white-loading)"/>
            <path className="tanuki-shape chin" d="m15.414 41.354 5.12 3.87 3.11 2.351a2.102 2.102 0 0 0 2.535 0l3.11-2.352 5.12-3.869-9.484-7.167-9.51 7.167Z" fill="var(--color-gray-loading)"/>
            <path className="tanuki-shape left-cheek" d="M10.019 22.923a22.86 22.86 0 0 0-9.117-4.1L.832 19A12.595 12.595 0 0 0 5.01 33.556l.026.02.057.044 10.32 7.734 9.491-7.167L10.02 22.923Z" fill="var(--color-white-loading)"/>
        </svg>);
    }
    return (<svg width={size} height={size} viewBox="0 0 32 32" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path fillRule="evenodd" clipRule="evenodd"
              d="M16 0C7.16 0 0 7.16 0 16C0 23.08 4.58 29.06 10.94 31.18C11.74 31.32 12.04 30.84 12.04 30.42C12.04 30.04 12.02 28.78 12.02 27.44C8 28.18 6.96 26.46 6.64 25.56C6.46 25.1 5.68 23.68 5 23.3C4.44 23 3.64 22.26 4.98 22.24C6.24 22.22 7.14 23.4 7.44 23.88C8.88 26.3 11.18 25.62 12.1 25.2C12.24 24.16 12.66 23.46 13.12 23.06C9.56 22.66 5.84 21.28 5.84 15.16C5.84 13.42 6.46 11.98 7.48 10.86C7.32 10.46 6.76 8.82 7.64 6.62C7.64 6.62 8.98 6.2 12.04 8.26C13.32 7.9 14.68 7.72 16.04 7.72C17.4 7.72 18.76 7.9 20.04 8.26C23.1 6.18 24.44 6.62 24.44 6.62C25.32 8.82 24.76 10.46 24.6 10.86C25.62 11.98 26.24 13.4 26.24 15.16C26.24 21.3 22.5 22.66 18.94 23.06C19.52 23.56 20.02 24.52 20.02 26.02C20.02 28.16 20 29.88 20 30.42C20 30.84 20.3 31.34 21.1 31.18C27.42 29.06 32 23.06 32 16C32 7.16 24.84 0 16 0V0Z"
              fill={fullColor}/>
    </svg>);
}

export function PlayPauseIcon({color = "white", isPlaying = false, size=25}) {
    const fullColor = `var(--color-${color})`;

    if (isPlaying)
        return (<svg width={size} height={size} viewBox="0 0 22 27" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M6 0C6.55228 0 7 0.447715 7 1V26C7 26.5523 6.55228 27 6 27H1C0.447715 27 4.02663e-09 26.5523 0 26V1C0 0.447715 0.447715 2.41595e-08 1 0H6ZM21 0C21.5523 0 22 0.447715 22 1V26C22 26.5523 21.5523 27 21 27H16C15.4477 27 15 26.5523 15 26V1C15 0.447715 15.4477 2.41595e-08 16 0H21Z" fill={fullColor} />
        </svg>);
    return (<svg width={size} height={size} viewBox="0 0 23 27" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path d="M21.6016 13.0769C21.9192 13.2721 21.9192 13.7333 21.6016 13.9285L1.26172 26.4294C0.928574 26.6342 0.499999 26.3947 0.499999 26.0037L0.5 1.00171C0.5 0.610677 0.928576 0.371177 1.26172 0.575927L21.6016 13.0769Z" fill={fullColor} />
    </svg>);
}

export function FullScreenIcon({color="white", iFullScreen=false, size=21}) {
    const fullColor = `var(--color-${color})`;

    if (iFullScreen)
        return (<svg width={size} height={size} xmlns="http://www.w3.org/2000/svg" viewBox="0 0 384 384">
            <path fill={fullColor} d="M0,356.3l6.35-12.73,90.21-90.56h-58.01c-.79,0-5.9-2.14-7.03-2.71-21.42-10.86-14.95-43.46,10.73-45.23,36.88-2.55,76.66,2,113.84,0,12.04.76,21.78,9.92,22.83,22.08,3.21,37.02-2.48,78.58,0,116.08-1.74,20.05-26.28,29.57-40.85,15.35-2.59-2.53-7.08-10.39-7.08-13.88v-57.26l-90.56,90.21-12.73,6.35h-7.49c-10.41-2.69-17.52-9.8-20.21-20.21v-7.49Z"/>
            <path fill={fullColor} d="M286.69,131.74h57.26c3.49,0,11.35,4.49,13.88,7.08,14.22,14.57,4.7,39.11-15.35,40.85-37.49-2.48-79.06,3.21-116.08,0-12.17-1.05-21.32-10.79-22.08-22.83,1.99-37.18-2.56-76.96,0-113.84,1.78-25.68,34.38-32.16,45.23-10.73.57,1.12,2.71,6.24,2.71,7.03v58.01L343.64,6.43c22.46-17.08,50,8.39,35.42,31.68l-92.38,93.63Z"/>
        </svg>);
    return (<svg width={size} height={size} xmlns="http://www.w3.org/2000/svg" viewBox="0 0 384 384">
        <path fill={fullColor} d="M0,243.38c4.31-18.28,26.86-26.36,40.82-12.75,2.59,2.53,7.08,10.38,7.08,13.87v57.22l91.8-91.12c23.44-17.19,51.17,10.98,33.48,34.21l-90.86,91.32h58c.95,0,6.26,2.31,7.5,2.97,16.7,8.93,16.19,33.6-.48,42.19l-7.4,2.7H20.21c-10.55-2.42-17.19-9.17-20.21-19.45v-121.17Z"/>
        <path fill={fullColor} d="M335.27,83.31l-91.37,90.81c-22.46,17.07-49.99-8.39-35.42-31.66l92.36-93.56h-57.25c-3.49,0-11.35-4.49-13.88-7.07-14.22-14.56-4.7-39.08,15.35-40.82,37.49,2.47,79.04-3.2,116.05,0,12.16,1.05,21.31,10.78,22.08,22.81-1.99,37.16,2.56,76.9,0,113.75-1.78,25.66-34.37,32.13-45.23,10.73-.57-1.12-2.71-6.24-2.71-7.02v-57.97Z"/>
    </svg>);
}

export function SubDelayIcon({direction="left", color="white", size=20}) {
    const fullColor = `var(--color-${color})`;

    if (direction === "left")
        return (<svg width={size} height={size} viewBox="0 0 30 27" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M30.0042 26.0039C30.0041 26.7859 29.147 27.2649 28.4808 26.8555L17.0042 19.8028L17.0042 25.6914C17.0042 26.5334 16.0269 26.9987 15.3733 26.4678L0.369424 14.2793C-0.123173 13.8792 -0.123101 13.1268 0.369424 12.7266L15.3733 0.538104C16.0269 0.00720582 17.0042 0.472477 17.0042 1.31447L17.0042 7.20314L28.4808 0.149431C29.147 -0.259802 30.0042 0.22 30.0042 1.00197L30.0042 26.0039Z" fill={fullColor}/>
        </svg>);
    return (<svg width={size} height={size} viewBox="0 0 30 27" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path d="M0 1.00122C0.00012226 0.219265 0.857203 -0.259781 1.52344 0.149658L13 7.20239V1.31372C13 0.471751 13.9773 0.00646258 14.6309 0.537354L29.6348 12.7258C30.1274 13.126 30.1273 13.8783 29.6348 14.2786L14.6309 26.467C13.9773 26.9979 13 26.5327 13 25.6907V19.802L1.52344 26.8557C0.857186 27.2649 -3.4181e-08 26.7851 0 26.0032V1.00122Z" fill={fullColor}/>
    </svg>);
}

export function CopyIcon({color="white", size=15}) {
    const fullColor = `var(--color-${color})`;

    return (<svg width={size} height={size} aria-hidden="true" focusable="false"
                 viewBox="0 0 16 16" fill="currentColor" display="inline-block"
                 overflow="visible">
        <path d="M0 6.75C0 5.784.784 5 1.75 5h1.5a.75.75 0 0 1 0 1.5h-1.5a.25.25 0 0 0-.25.25v7.5c0 .138.112.25.25.25h7.5a.25.25 0 0 0 .25-.25v-1.5a.75.75 0 0 1 1.5 0v1.5A1.75 1.75 0 0 1 9.25 16h-7.5A1.75 1.75 0 0 1 0 14.25Z" fill={fullColor}/>
        <path d="M5 1.75C5 .784 5.784 0 6.75 0h7.5C15.216 0 16 .784 16 1.75v7.5A1.75 1.75 0 0 1 14.25 11h-7.5A1.75 1.75 0 0 1 5 9.25Zm1.75-.25a.25.25 0 0 0-.25.25v7.5c0 .138.112.25.25.25h7.5a.25.25 0 0 0 .25-.25v-7.5a.25.25 0 0 0-.25-.25Z" fill={fullColor}/>
    </svg>);
}
