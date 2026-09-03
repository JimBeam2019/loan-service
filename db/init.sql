
--
-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;


--
-- Name: EXTENSION "uuid-ossp"; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION "uuid-ossp" IS 'generate universally unique identifiers (UUIDs)';


--
-- Name: loan_state; Type: TYPE; Schema: public; Owner: -
--
DO $$
BEGIN
    CREATE TYPE loan_state AS ENUM (
        'proposed',
        'approved',
        'invested',
        'disbursed'
    );
EXCEPTION
    WHEN duplicate_object THEN NULL;
END
$$;

--
-- Name: loans; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS loans (
    id uuid DEFAULT uuid_generate_v4() PRIMARY KEY NOT NULL,
    borrower_id uuid NOT NULL,
    principal_amount numeric(24,10) NOT NULL,
    rate numeric(24,10) NOT NULL,
    return_of_investment numeric(24,10) NOT NULL,
    agreement_letter_url varchar(250),
    state loan_state DEFAULT 'proposed'::loan_state NOT NULL,
    proposed_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,

    CONSTRAINT chk_principal_positive CHECK (principal_amount > 0),
    CONSTRAINT chk_rate_nonnegative CHECK (rate >= 0),
    CONSTRAINT chk_roi_nonnegative CHECK (return_of_investment >= 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_active_loan_request
ON loans (
    borrower_id,
    principal_amount,
    rate,
    return_of_investment
)
WHERE state IN ('proposed'::loan_state, 'approved'::loan_state, 'invested'::loan_state);

--
-- Name: approvals; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS approvals (
    id uuid DEFAULT uuid_generate_v4() PRIMARY KEY NOT NULL,
    loan_id uuid NOT NULL UNIQUE,
    field_validator_employee_id uuid NOT NULL,
    visit_proof_url varchar(250) NOT NULL,
    approved_at timestamp with time zone NOT NULL,

    CONSTRAINT fk_approval_loan FOREIGN KEY (loan_id) REFERENCES loans(id)
);

--
-- Name: investments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS investments (
    id uuid DEFAULT uuid_generate_v4() PRIMARY KEY NOT NULL,
    loan_id uuid NOT NULL,
    investor_id uuid NOT NULL,
    amount numeric(24,10) NOT NULL,
    agreement_letter_url varchar(250) NOT NULL,
    invested_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,

    CONSTRAINT fk_investment_loan FOREIGN KEY (loan_id) REFERENCES loans(id),
    CONSTRAINT chk_investment_amount_positive CHECK (amount > 0),
    CONSTRAINT uq_loan_investor UNIQUE (loan_id, investor_id)
);

--
-- Name: disbursements; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE IF NOT EXISTS disbursements (
    id uuid DEFAULT uuid_generate_v4() PRIMARY KEY NOT NULL,
    loan_id uuid NOT NULL UNIQUE,
    signed_agreement_url varchar(250) NOT NULL,
    field_officer_employee_id uuid NOT NULL,
    disbursed_at timestamp with time zone NOT NULL,

    CONSTRAINT fk_disbursement_loan FOREIGN KEY (loan_id) REFERENCES loans(id)
);