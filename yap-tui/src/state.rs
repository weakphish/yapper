#[derive(Debug, Default, PartialEq, Eq)]
pub enum ViewState {
    #[default]
    DailyPage,
    TagPage,
    Exit,
}
