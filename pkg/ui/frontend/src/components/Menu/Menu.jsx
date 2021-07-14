import React from 'react';
import { makeStyles } from '@material-ui/core/styles';
import Drawer from '@material-ui/core/Drawer';
import CssBaseline from '@material-ui/core/CssBaseline';
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import List from '@material-ui/core/List';
import Typography from '@material-ui/core/Typography';
import Divider from '@material-ui/core/Divider';
import ListItem from '@material-ui/core/ListItem';
import ListItemIcon from '@material-ui/core/ListItemIcon';
import ListItemText from '@material-ui/core/ListItemText';
import { blue } from '@material-ui/core/colors';
import Grid from '@material-ui/core/Grid';
import { Link } from 'react-router-dom';
import NoteAddIcon from '@material-ui/icons/NoteAdd';
import Person from '@material-ui/icons/Person';
import WatchLaterIcon from '@material-ui/icons/WatchLater';

const drawerWidth = 240;
const useStyles = makeStyles(theme => ({
    root: {
        display: 'flex',
    },
    appBar: {
        width: `calc(100% - ${drawerWidth}px)`,
        marginLeft: drawerWidth,
    },
    drawer: {
        width: drawerWidth,
        flexShrink: 0,
    },
    drawerPaper: {
        width: drawerWidth,
    },
    // necessary for content to be below app bar
    toolbar: theme.mixins.toolbar,
    content: {
        flexGrow: 1,
        backgroundColor: theme.palette.background.default,
        padding: theme.spacing(3),
    },
}));

export default function PermanentDrawerLeft() {
    const classes = useStyles();
    const color = 'primary';
    const variant = 'h6';
    return (
        <div className={classes.root}>
            <CssBaseline />
            <AppBar position="fixed" className={classes.appBar}>
                <Toolbar></Toolbar>
            </AppBar>
            <Drawer
                className={classes.drawer}
                variant="permanent"
                classes={{
                    paper: classes.drawerPaper,
                }}
                anchor="left"
            >
                <div className={classes.toolbar}>
                    <Grid
                        container
                        spacing={0}
                        direction="column"
                        alignItems="center"
                        justify="center"
                    >
                        <img
                            src={process.env.PUBLIC_URL + '/logo_menu.png'}
                            align="center"
                            height="60"
                            width="100"
                        />
                    </Grid>
                </div>
                <Divider />
                <List component="div" disablePadding>
                    <ListItem button className={classes.nested} component={Link} to="/">
                        <ListItemIcon>
                            <Person style={{ color: blue[500] }} fontSize="large" />
                        </ListItemIcon>
                        <ListItemText>
                            <Typography variant={variant} color={color}>
                                Welcome
                            </Typography>
                        </ListItemText>
                    </ListItem>

                    <ListItem button className={classes.nested} component={Link} to="/hp">
                        <ListItemIcon>
                            <NoteAddIcon style={{ color: blue[500] }} fontSize="large" />
                        </ListItemIcon>
                        <ListItemText>
                            <Typography variant={variant} color={color}>
                                Submit
                            </Typography>
                        </ListItemText>
                    </ListItem>
                    <ListItem button className={classes.nested} component={Link} to="/hp_monitor">
                        <ListItemIcon>
                            <WatchLaterIcon style={{ color: blue[500] }} fontSize="large" />
                        </ListItemIcon>
                        <ListItemText>
                            <Typography variant={variant} color={color}>
                                Monitor
                            </Typography>
                        </ListItemText>
                    </ListItem>
                </List>
                <Divider />
            </Drawer>
        </div>
    );
}
