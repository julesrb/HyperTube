type tDataErrorResponse = {
    error: {
        code: string
        message?: string
        fields?: Record<string, {message: string}>
    }
};

export class ApiError extends Error {
    status: number;
    data: tDataErrorResponse;
    notificationMsg: string;

    constructor(status: number, data: tDataErrorResponse) {
        super(data?.error?.message ?? "Unknown error occurred.");

        this.name = "ApiError";
        this.status = status;
        this.data = data;
        this.notificationMsg = `${status} - ${this.message}`;
        Object.setPrototypeOf(this, ApiError.prototype);
    }
}
