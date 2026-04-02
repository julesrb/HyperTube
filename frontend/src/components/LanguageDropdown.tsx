import styles from "./LanguageDropdown.module.css";

export default function LanguageDropdown() {
    return (<div className={styles.dropdown}>
            <button className="clean-btn">English</button>
            <button className="clean-btn">French</button>
            <button className="clean-btn">Spanish</button>
    </div>)
}
